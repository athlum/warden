package main

import (
	"github.com/athlum/warden/warden"
	"github.com/athlum/warden/warden/appconfig"
	"github.com/athlum/warden/warden/backends"
	"github.com/athlum/warden/warden/utils"
	"github.com/athlum/warden/warden/watcher"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var config *warden.AgentConfig

func init() {
	utils.WarningAsRoot()
	config = warden.AgentSettings()
	if err := backends.NewZKConn(config.ZKHost, config.ZKAuth, warden.DefaultZKConnectionTime); err != nil {
		log.Fatal(err)
	}
	if err := appconfig.NewConfigCenter(config.AppHome); err != nil {
		log.Fatal(err)
	}
}

func main() {
	daemonWatcher, daemonError := watcher.NewDaemonWatcher(config.DockerSocket)
	if daemonError != nil {
		log.Fatal(daemonError)
	}

	syncErr := daemonWatcher.SyncContainers(&Handler{})
	if syncErr != nil {
		log.Fatal(syncErr)
	}

	watcherExit := make(chan string)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		daemonWatcher.Watch(&Handler{}, &watcherExit)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGHUP)
	s := <-c
	if s != syscall.SIGHUP {
		watcherExit <- "quit"
		signal.Stop(c)
		wg.Wait()
		os.Exit(0)
	} else {
		warden.ReloadAgentSettings()
	}
}
