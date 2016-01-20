package appconfig

import (
	"errors"
	"fmt"
	"github.com/athlum/warden/warden"
	docker "github.com/fsouza/go-dockerclient"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var HTTPRequestError = errors.New("Request error.")

type HTTPCenter struct {
	*BaseCenter
}

func (config *HTTPCenter) Exists(appId string) (bool, error) {
	appPath := config.AppPath(appId)
	log.Println(appPath)
	if resp, err := http.Get(appPath); err != nil {
		return false, err
	} else {
		if resp != nil && resp.StatusCode < 400 {
			return true, nil
		}

		return false, HTTPRequestError
	}
}

func (config *HTTPCenter) Download(src string, dest string) error {
	log.Println(src)
	resp, err := http.Get(src)
	if err != nil {
		return err
	} else if resp.StatusCode >= 400 {
		return HTTPRequestError
	}
	defer resp.Body.Close()
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return readErr
	}
	err = ioutil.WriteFile(dest, body, os.FileMode(uint32(0550)))
	return err
}

func (config *HTTPCenter) GetFile(src string, dest string) error {
	if err := config.Download(src, dest); err != nil {
		return err
	}
	return nil
}

func (config *HTTPCenter) GetAppMeta(appId string) (*warden.AppMeta, error) {
	appMeta := &warden.AppMeta{}

	var metaFile = fmt.Sprintf("%s/%s-meta.json", warden.HTTPCACHEDIR, appId)
	metaUrl := fmt.Sprintf("%s/meta.json", config.AppPath(appId))
	if err := config.GetFile(metaUrl, metaFile); err != nil {
		log.Printf("Error: %s", err)
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

func (config *HTTPCenter) GetApp(appId string) (*warden.Application, error) {
	if exists, _ := config.Exists(appId); !exists {
		return nil, AppNotExist
	}

	appMeta, err := config.GetAppMeta(appId)
	if err != nil {
		return nil, MetaError
	}

	var validateShell = fmt.Sprintf("%s/%s-validate.sh", warden.HTTPCACHEDIR, appId)
	shellUrl := fmt.Sprintf("%s/validate.sh", config.AppPath(appId))
	if err := config.GetFile(shellUrl, validateShell); err != nil {
		log.Printf("Error: %s", err)
		return nil, ValidateShellNotExist
	}

	var registerShell = fmt.Sprintf("%s/%s-register.sh", warden.HTTPCACHEDIR, appId)
	shellUrl = fmt.Sprintf("%s/register.sh", config.AppPath(appId))
	if err := config.GetFile(shellUrl, registerShell); err != nil {
		log.Printf("Error: %s", err)
		return nil, RegisterShellNotExist
	}

	return &warden.Application{appMeta, appId, validateShell, registerShell}, nil
}

func (config *HTTPCenter) GetDaemonApp(container *docker.Container) (*warden.DaemonApplication, error) {
	ipAddress, hostName, envKv := config.Inspect(container)

	appId := envKv["APPID"]

	application, err := config.GetApp(appId)
	if err != nil {
		return nil, err
	}
	return &warden.DaemonApplication{application, hostName, fmt.Sprintf("%s:%s", ipAddress, envKv["PORT"]), ipAddress, envKv["PORT"]}, nil
}
