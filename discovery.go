package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/Jeffail/gabs"
	"github.com/labstack/gommon/log"
)

func DiscoverBalenaDevices() ([]*url.URL, error) {
	apiKey := os.Getenv("API_KEY") // Done in balena panel
	if apiKey == "" {
		panic("Api key not found, cannot contact balena api")
	}

	client := http.Client{}
	req, err := http.NewRequest("GET", "https://api.balena-cloud.com/v6/device?$filter=belongs_to__application%20eq%20'1869612'", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", apiKey))
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	jsonParsed, err := gabs.ParseJSON(respBody)
	if err != nil {
		return nil, err
	}

	children, err := jsonParsed.S("d").Children()
	if err != nil {
		return nil, err
	}
	devices := []*url.URL{}
	for _, device := range children {
		fmt.Println(device.Data().(map[string]interface{})["ip_address"], "found in API")
		u, err := url.Parse(fmt.Sprintf("http://%s:80", device.Data().(map[string]interface{})["ip_address"].(string)))
		if err != nil {
			log.Error("Error parsing IP from API:")
			log.Error(err)
		}
		devices = append(devices, u)
	}
	return devices, nil
}
