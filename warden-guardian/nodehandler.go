package main

import (
	"encoding/json"
	"git.elenet.me/qi.feng/warden/appconfig"
	"log"
)

type Handler struct {
}

func (handler *Handler) Register(appId string, nodes interface{}) error {
	configCenter := appconfig.AppConfigCenter()
	app, appErr := configCenter.GetApp(appId)
	if appErr != nil {
		return appErr
	}

	data, err := json.Marshal(nodes)
	if err != nil {
		log.Println("error:", err)
	}
	if err := app.Register(string(data)); err != nil {
		return err
	}
	return nil
}
