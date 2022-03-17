package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	devopsv1 "k8s-crd-caddy/api/v1"
	networkingv1 "k8s.io/api/networking/v1beta1"
	"net/http"
	"time"
)

type CaddyRoute struct {
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
	body := map[string]interface{}{
		"match": []map[string]interface{}{
			{
				"path": []string{app.Spec.Path},
			},
		},
		"handle": []map[string]interface{}{
			{
				"body":    app.Spec.Content,
				"handler": "static_response",
			},
		},
	}
	bodyByte, err := json.Marshal(body)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(bodyByte)
	req, err := http.NewRequest(http.MethodPost, "http://caddy-controller.codepy.net/config/apps/http/servers/srv0/routes", reader)
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

func DeleteCaddyRoute(ingress networkingv1.IngressSpec) error {
	client := http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, "http://caddy-controller.codepy.net/config/apps/http/servers/srv0/routes", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var routes []CaddyRoute
	if err := json.Unmarshal(respByte, &routes); err != nil {
		return err
	}
	path := ingress.Rules[0].HTTP.Paths[0].Path
	for idx, route := range routes {
		if path == route.Match[0].Path[0] {
			delReq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://caddy-controller.codepy.net/config/apps/http/servers/srv0/routes/%d", idx), nil)
			if err != nil {
				return err
			}
			_, err = client.Do(delReq)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func DeleteCaddyRouteInstance(app *devopsv1.Static) error {
	path := app.Spec.Path
	client := http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, "http://caddy-controller.codepy.net/config/apps/http/servers/srv0/routes", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var routes []CaddyRoute
	if err := json.Unmarshal(respByte, &routes); err != nil {
		return err
	}
	for idx, route := range routes {
		if path == route.Match[0].Path[0] {
			delReq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://caddy-controller.codepy.net/config/apps/http/servers/srv0/routes/%d", idx), nil)
			if err != nil {
				return err
			}
			_, err = client.Do(delReq)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
