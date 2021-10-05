package main

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	Server *echo.Echo
	// Old url list to compare to
	urlList    []*url.URL = []*url.URL{} // Start empty
	cidrPrefix            = "192.168.1"
	targets    []*middleware.ProxyTarget
	ticker     *time.Ticker
	rrb        middleware.ProxyBalancer
)

func ScanForUpstreams() {
	foundURLs := []*url.URL{}
	for i := 0; i < 255; i++ {
		// fmt.Println("checking", fmt.Sprintf("%s.%d", cidrPrefix, i))
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(fmt.Sprintf("%s.%d", cidrPrefix, i), "80"), time.Millisecond*200)
		if err != nil {
			// fmt.Println("Failed to connect to", net.JoinHostPort(fmt.Sprintf("%s.%d", cidrPrefix, i), "80"))
			// fmt.Println(err)
		}
		if conn != nil {
			// fmt.Println("Connected to device", fmt.Sprintf("%s.%d", cidrPrefix, i))
			conn.Close()
			u, err := url.Parse(fmt.Sprintf("http://%s", net.JoinHostPort(fmt.Sprintf("%s.%d", cidrPrefix, i), "80")))
			if err != nil {
				panic(err)
			}
			foundURLs = append(foundURLs, u)
		}
		time.Sleep(time.Millisecond * 100)
	}

	// Compare url lists to see what to add and what to drop
	newURLs := []*url.URL{}
	for _, i := range foundURLs {
		found := false
		for _, j := range urlList {
			if i == j {
				found = true
			}
		}
		// New upstream
		if !found {
			fmt.Println("Found a new upstream:", i.String())
			newURLs = append(newURLs, i)
		}
	}

	deadURLs := []*url.URL{}
	for _, i := range urlList {
		found := false
		for _, j := range foundURLs {
			if i == j {
				found = true
			}
		}
		// New upstream
		if !found {
			fmt.Println("Upstream died!", i.String())
			deadURLs = append(deadURLs, i)
		}
	}

	// Add new
	for _, i := range newURLs {
		rrb.AddTarget(&middleware.ProxyTarget{
			URL: i,
		})
	}
	// Remove dead
	for _, i := range deadURLs {
		rrb.RemoveTarget(i.String())
	}

	// Copy over
	urlList = foundURLs
}

func StartServer() {
	Server = echo.New()
	Server.HideBanner = true

	rrb = middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{}) // No targets
	ScanForUpstreams()
	Server.Use(middleware.Proxy(rrb))

	Server.Start(":80")
}

func StartScanTicke() {
	ticker = time.NewTicker(5 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				fmt.Println("Ticker exiting")
				return
			case <-ticker.C:
				ScanForUpstreams()
			}
		}
	}()
}
