package handlers

import (
	docker "github.com/fsouza/go-dockerclient"
)

type DaemonHandler interface {
	OnStart(container *docker.Container, exit *chan string) error
	OnDie(event *docker.APIEvents) error
}
