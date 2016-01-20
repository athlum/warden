package warden

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

const (
	HTTPCACHEDIR       = "/opt/warden"
	AgentConfigFile    = "/etc/warden/warden-agent.json"
	GuardianConfigFile = "/etc/warden/warden-guardian.json"
)

var (
	DefaultZKConnectionTime    = time.Second * 30
	agentSettings              = AgentConfig{}
	guardianSettings           = GuardianConfig{}
	DefaultHealthCheckInterval = 5
	DefaultHealthCheckRetry    = 3
	ConfigLock                 = new(sync.RWMutex)
)

type AgentConfig struct {
	ZKHost       []string `json:"zkhost"`
	ZKAuth       string   `json:"zkauth"`
	AppHome      string   `json:"apphome"`
	DockerSocket string   `json:"dockersocket"`
}

type GuardianConfig struct {
	ZKHost       []string `json:"zkhost"`
	ZKAuth       string   `json:"zkauth"`
	AppHome      string   `json:"apphome"`
	Applications []string `json:"applications"`
}

func loadAgentConfig(config *AgentConfig, configfile string) error {
	configBody, readErr := ioutil.ReadFile(configfile)
	if readErr != nil {
		return readErr
	}
	ConfigLock.Lock()
	defer ConfigLock.Unlock()
	err := json.Unmarshal(configBody, config)
	if err != nil {
		return err
	}

	log.Printf("%+v", config)
	return nil
}

func loadGuardianConfig(config *GuardianConfig, configfile string) error {
	configBody, readErr := ioutil.ReadFile(configfile)
	if readErr != nil {
		return readErr
	}
	ConfigLock.Lock()
	defer ConfigLock.Unlock()
	err := json.Unmarshal(configBody, config)
	if err != nil {
		return err
	}

	log.Printf("%+v", config)
	return nil
}

func init() {
	if _, err := os.Stat(AgentConfigFile); err == nil {
		loadAgentConfig(&agentSettings, AgentConfigFile)
	}
	if _, err := os.Stat(GuardianConfigFile); err == nil {
		loadGuardianConfig(&guardianSettings, GuardianConfigFile)
	}
}

func AgentSettings() *AgentConfig {
	return &agentSettings
}

func ReloadAgentSettings() error {
	err := loadAgentConfig(&agentSettings, AgentConfigFile)
	return err
}

func ReloadGuardianSettings() error {
	err := loadGuardianConfig(&guardianSettings, GuardianConfigFile)
	return err
}

func GuardianSettings() *GuardianConfig {
	return &guardianSettings
}
