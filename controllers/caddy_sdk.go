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
	reader := new(bytes.Buffer)
	if err := json.NewEncoder(reader).Encode(body); err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://caddy-controller.%s:2019/config/apps/http/servers/srv0/routes", app.Namespace), reader)
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

func DeleteCaddyRoute(id, namespace string) error {
	client := http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://caddy-controller.%s:2019/id/%s", namespace, id), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if _, err := client.Do(req); err != nil {
		return err
	}
	return nil
}
