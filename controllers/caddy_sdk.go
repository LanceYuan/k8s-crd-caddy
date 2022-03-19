package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	devopsv1 "k8s-crd-caddy/api/v1"
	"net/http"
	"time"
)

type CaddyRoute struct {
	Id     string             `json:"@id"`
	Match  []CaddyRouteMatch  `json:"match"`
	Handle []CaddyRouteHandle `json:"handle"`
}

type CaddyRouteMatch struct {
	Path []string `json:"path"`
}

type CaddyRouteHandle struct {
	Body    string `json:"body"`
	Handler string `json:"handler"`
}

func AddCaddyRoute(app *devopsv1.Static) error {
	client := http.Client{Timeout: 10 * time.Second}
	body := CaddyRoute{
		Id: app.Name,
		Match: []CaddyRouteMatch{
			{
				Path: []string{app.Spec.Path},
			},
		},
		Handle: []CaddyRouteHandle{
			{
				Body:    app.Spec.Content,
				Handler: "static_response",
			},
		},
	}
	bodyByte, err := json.Marshal(body)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(bodyByte)
	req, err := http.NewRequest(http.MethodPost, "http://caddy-controller/config/apps/http/servers/srv0/routes", reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func DeleteCaddyRoute(id string) error {
	client := http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://caddy-controller/id/%s", id), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if _, err := client.Do(req); err != nil {
		return err
	}
	return nil
}
