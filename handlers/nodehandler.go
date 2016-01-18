package handlers

type NodeHandler interface {
	Register(appId string, nodes interface{}) error
}
