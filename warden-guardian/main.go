package main

import (
	"github.com/athlum/warden"
	"github.com/athlum/warden/appconfig"
	"github.com/athlum/warden/backends"
	"github.com/athlum/warden/utils"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var config *warden.GuardianConfig

func init() {
	utils.WarningAsRoot()
	config = warden.GuardianSettings()
	if err := backends.NewZKConn(config.ZKHost, config.ZKAuth, warden.DefaultZKConnectionTime); err != nil {
		log.Fatal(err)
	}
	if err := appconfig.NewConfigCenter(config.AppHome); err != nil {
		log.Fatal(err)
	}
}

func main() {
	wg := sync.WaitGroup{}
	client := backends.ZKClient()
	quitChan := make(map[string]*chan string)
	for _, appId := range config.Applications {
		quit := make(chan string)
		wg.Add(1)
		go func(appId string) {
			defer wg.Done()
			client.Watch(appId, &Handler{}, &quit)
		}(appId)
		quitChan[appId] = &quit
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGHUP)
	s := <-c
	if s != syscall.SIGHUP {
		for _, quit := range quitChan {
			*quit <- "quit"
		}
		signal.Stop(c)
		wg.Wait()
		os.Exit(0)
	} else {
		warden.ReloadGuardianSettings()
	}
}
