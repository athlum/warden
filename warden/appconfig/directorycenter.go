package appconfig

import (
	"fmt"
	"github.com/athlum/warden/warden"
	docker "github.com/fsouza/go-dockerclient"
	"io/ioutil"
	"log"
	"os"
)

type DirectoryCenter struct {
	*BaseCenter
}

func (config *DirectoryCenter) Exists(appId string) (bool, error) {
	appPath := config.AppPath(appId)
	log.Println(appPath)
	if _, err := os.Stat(appPath); err != nil {
		return false, nil
	}
	return true, nil
}

func (config *DirectoryCenter) GetAppMeta(appId string) (*warden.AppMeta, error) {
	appMeta := &warden.AppMeta{}

	var metaFile = fmt.Sprintf("%s/meta.json", config.AppPath(appId))
	if _, err := os.Stat(metaFile); err != nil {
		return appMeta, nil
	}

	body, readErr := ioutil.ReadFile(metaFile)
	if readErr != nil {
		log.Println("Read meta failed:", readErr)
		return appMeta, nil
	}

	appMeta, loadErr := config.LoadAppMeta(body)
	if loadErr != nil {
		return nil, loadErr
	}

	return appMeta, nil
}

func (config *DirectoryCenter) GetApp(appId string) (*warden.Application, error) {
	if exists, _ := config.Exists(appId); !exists {
		return nil, AppNotExist
	}

	appMeta, err := config.GetAppMeta(appId)
	if err != nil {
		return nil, MetaError
	}

	var validateShell = fmt.Sprintf("%s/validate.sh", config.AppPath(appId))
	if _, err := os.Stat(validateShell); err != nil {
		return nil, ValidateShellNotExist
	}

	var registerShell = fmt.Sprintf("%s/register.sh", config.AppPath(appId))
	if _, err := os.Stat(registerShell); err != nil {
		return nil, RegisterShellNotExist
	}

	return &warden.Application{appMeta, appId, validateShell, registerShell}, nil
}

func (config *DirectoryCenter) GetDaemonApp(container *docker.Container) (*warden.DaemonApplication, error) {
	ipAddress, hostName, envKv := config.Inspect(container)
	appId := envKv["APPID"]

	application, err := config.GetApp(appId)
	if err != nil {
		return nil, err
	}

	return &warden.DaemonApplication{application, hostName, fmt.Sprintf("%s:%s", ipAddress, envKv["PORT"]), ipAddress, envKv["PORT"]}, nil
}
