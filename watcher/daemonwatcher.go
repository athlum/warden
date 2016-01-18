package watcher

import (
	"github.com/athlum/warden/handlers"
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"time"
)

type DaemonWatcher struct {
	*docker.Client
	managedContainer map[string]*chan string
}

func NewDaemonWatcher(endpoint string) (*DaemonWatcher, error) {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		return nil, err
	}
	log.Println("Docker daemon connected.")
	return &DaemonWatcher{client, make(map[string]*chan string)}, nil
}

func (watcher *DaemonWatcher) SyncContainers(handler handlers.DaemonHandler) error {
	containers, err := watcher.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return err
	}
	for _, c := range containers {
		exit := make(chan string)
		container, err := watcher.InspectContainer(c.ID)
		if err != nil {
			log.Printf("Error: Inspect %s failed.", c.ID)
			break
		}
		if err := handler.OnStart(container, &exit); err != nil {
			log.Println("Error:", err)
			continue
		}
		watcher.managedContainer[c.ID] = &exit
	}

	return nil
}

func (watcher *DaemonWatcher) removeContainerCh(key string) {
	select {
	case *watcher.managedContainer[key] <- "quit":
		log.Printf("Send 'quit' signal to container %s health checker.", key)
	case <-time.After(time.Microsecond * 100):
		log.Printf("Health checker for container %s already exited.", key)
	}
	delete(watcher.managedContainer, key)
}

func (watcher *DaemonWatcher) Watch(handler handlers.DaemonHandler, watcherExit *chan string) {
	listener := make(chan *docker.APIEvents)
	err := watcher.AddEventListener(listener)
	if err != nil {
		return
	}
	for {
		select {
		case event := <-listener:
			if event.Status == "start" {
				exit := make(chan string)
				container, err := watcher.InspectContainer(event.ID)
				if err != nil {
					log.Printf("Error: Inspect %s failed.", event.ID)
					break
				}
				if err := handler.OnStart(container, &exit); err != nil {
					log.Println("Error:", err)
					break
				}
				watcher.managedContainer[event.ID] = &exit
			} else if event.Status == "die" {
				if _, ok := watcher.managedContainer[event.ID]; ok {
					if err := handler.OnDie(event); err != nil {
						break
					}
					watcher.removeContainerCh(event.ID)
				}
			}
		case <-*watcherExit:
			for key, _ := range watcher.managedContainer {
				watcher.removeContainerCh(key)
			}
			return
		}
	}
}
