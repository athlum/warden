package main

import (
	"git.elenet.me/qi.feng/warden/appconfig"
	"git.elenet.me/qi.feng/warden/backends"
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"time"
)

type Handler struct{}

func (handler *Handler) OnStart(container *docker.Container, exit *chan string) error {
	configCenter := appconfig.AppConfigCenter()
	app, appErr := configCenter.GetDaemonApp(container)
	if appErr != nil {
		return appErr
	}

	count := 1
	var interval time.Duration
	if app.StartAwait != 0 {
		count = 3
		interval = time.Duration(app.StartAwait/count) * time.Second
	}
	log.Printf("Start await: total %ds, retry %d, interval %ds", app.StartAwait, count, interval/time.Second)

	var valiErr error
	for i := 0; i < count; i++ {
		time.Sleep(interval)
		valiErr = app.Validate()
		if valiErr == nil {
			break
		}
	}
	if valiErr != nil {
		return valiErr
	}

	client := backends.ZKClient()
	if err := client.NewNode(app.NodePath(), app.Node); err != nil {
		return err
	}
	go app.WatchTillDie(exit)
	return nil
}

func (handler *Handler) OnDie(event *docker.APIEvents) error {
	return nil
}
