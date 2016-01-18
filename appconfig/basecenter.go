package appconfig

import (
	"encoding/json"
	"fmt"
	"git.elenet.me/qi.feng/warden"
	"git.elenet.me/qi.feng/warden/utils"
	docker "github.com/fsouza/go-dockerclient"
	"strings"
)

type BaseCenter struct {
	AppCenter string
}

func (config *BaseCenter) LoadAppMeta(body []byte) (*warden.AppMeta, error) {
	appMeta := warden.AppMeta{}
	err := json.Unmarshal(body, &appMeta)
	if err != nil {
		return nil, err
	}
	return &appMeta, nil
}

func (config *BaseCenter) AppPath(appId string) string {
	if strings.HasSuffix(config.AppCenter, "/") {
		return fmt.Sprintf("%s%s", config.AppCenter, appId)
	}
	return fmt.Sprintf("%s/%s", config.AppCenter, appId)
}

func (config *BaseCenter) Inspect(container *docker.Container) (ipAddress string, hostName string, envKv map[string]string) {
	ipAddress = container.NetworkSettings.IPAddress
	hostName = container.Config.Hostname
	envKv = utils.EnvToMap(container.Config.Env)
	return
}
