package backends

import (
	"git.elenet.me/qi.feng/warden/handlers"
)

type AppNode struct {
	HostName string `json:"hostname"`
	Node     string `json:"node"`
}

type backends interface {
	Watch(appId string, handler handlers.NodeHandler)
	NewNode(path string, value string) error
	RemoveNode(path string) error
}
