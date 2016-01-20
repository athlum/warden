package appconfig

import (
	"errors"
	// "fmt"
	"github.com/athlum/warden/warden"
	docker "github.com/fsouza/go-dockerclient"
	"os"
	"strings"
	"sync"
)

var (
	AppNotExist           = errors.New("App does not exist.")
	MetaError             = errors.New("Load app meta error.")
	ValidateShellNotExist = errors.New("Shell for validation does not exist.")
	RegisterShellNotExist = errors.New("Shell for registration does not exist.")
	ConfigCenterLock      = new(sync.RWMutex)
)

var appConfigCenter ConfigCenter

type ConfigCenter interface {
	Exists(appId string) (bool, error)
	GetAppMeta(appId string) (*warden.AppMeta, error)
	GetApp(appId string) (*warden.Application, error)
	GetDaemonApp(container *docker.Container) (*warden.DaemonApplication, error)
}

func NewConfigCenter(appHome string) error {
	ConfigCenterLock.Lock()
	defer ConfigCenterLock.Unlock()
	bc := &BaseCenter{appHome}
	switch {
	case strings.HasPrefix(appHome, "/"):
		appConfigCenter = &DirectoryCenter{bc}
	case strings.HasPrefix(appHome, "http"):
		appConfigCenter = &HTTPCenter{bc}
		os.Mkdir(warden.HTTPCACHEDIR, os.FileMode(uint32(0660)))
	default:
		appConfigCenter = &DirectoryCenter{bc}
	}
	return nil
}

func AppConfigCenter() ConfigCenter {
	return appConfigCenter
}
