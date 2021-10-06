package main

import (
	"fmt"
	"net"
	"net/url"

	"github.com/hashicorp/mdns"
	"github.com/labstack/gommon/log"
)

func DiscoverMDNSDevices() []*url.URL {
	foundURLs := []*url.URL{}
	// Find them
	entriesCh := make(chan *mdns.ServiceEntry, 256)
	go func() {
		for entry := range entriesCh {
			if entry.Info == "Balena cluster service" {
				fmt.Printf("Got new entry: %v\n", entry)
				u, err := url.Parse(fmt.Sprintf("http://%s", net.JoinHostPort(entry.AddrV4.String(), "80")))
				if err != nil {
					log.Errorf("Error found: %v", err)
				}
				foundURLs = append(foundURLs, u)
			}
		}
	}()

	// Start the lookup
	err := mdns.Lookup("_balena-worker._tcp", entriesCh)
	if err != nil {
		panic(err)
	}
	close(entriesCh)
	return foundURLs
}
