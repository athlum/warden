package warden

import (
	"errors"
	"fmt"
	"github.com/athlum/warden/backends"
	"log"
	"os/exec"
	"time"
)

var (
	ValidationFailed = errors.New("Health Check Failed.")
	RegistFailed     = errors.New("App Registration Failed.")
)

type AppMeta struct {
	StartAwait int `json:"startawait"`
	HCInterval int `json:"hcinterval"`
	HCRetry    int `json:"hcretry"`
}

func (app *AppMeta) HealthCheckInterval() int {
	if app.HCInterval == 0 {
		return DefaultHealthCheckInterval
	}
	return app.HCInterval
}

func (app *AppMeta) HealthCheckRetry() int {
	if app.HCRetry == 0 {
		return DefaultHealthCheckRetry
	}
	return app.HCRetry
}

type Application struct {
	*AppMeta
	AppId         string
	ValidateShell string
	RegisterShell string
}

func (app *Application) NodePath() string {
	return fmt.Sprintf("/app/%s", app.AppId)
}

func (app *Application) Register(data string) error {
	p := exec.Command(app.RegisterShell, data)
	log.Println("Registration:", p.Args)
	if err := p.Run(); err == nil {
		return nil
	} else {
		log.Println("Error:", err)
	}
	return RegistFailed
}

type DaemonApplication struct {
	*Application
	HostName  string
	Node      string
	IpAddress string
	Port      string
}

func (app *DaemonApplication) NodePath() string {
	return fmt.Sprintf("/app/%s/%s", app.AppId, app.HostName)
}

func (app *DaemonApplication) Validate() error {
	p := exec.Command(app.ValidateShell, app.IpAddress, app.Port)
	log.Println("Validation:", p.Args)
	if err := p.Run(); err == nil {
		return nil
	} else {
		log.Println("Error:", err)
	}
	return ValidationFailed
}

func (app *DaemonApplication) WatchTillDie(exit *chan string) {
	client := backends.ZKClient()
	interval := time.Duration(app.HealthCheckInterval()) * time.Second
	retry := 0
	for {
		select {
		case <-*exit:
			log.Printf("Container %s dead.", app.IpAddress)
			if err := client.RemoveNode(app.NodePath()); err != nil {
				log.Println("Error on quit:", err)
				continue
			}
			return
		case <-time.After(interval):
			if err := app.Validate(); err != nil && retry >= app.HealthCheckRetry() {
				if err := client.RemoveNode(app.NodePath()); err != nil {
					log.Println("Error on quit:", err)
					continue
				}
				return
			} else if err != nil {
				retry += 1
				log.Printf("Retried %d on %s", retry, app.IpAddress)
			}
		}
	}
}
