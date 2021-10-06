package main

import (
	"fmt"
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
	fmt.Println("Scanning for devices...")
	foundURLs, err := DiscoverBalenaDevices()
	if err != nil {
		panic(err)
	}

	// Compare url lists to see what to add and what to drop
	newURLs := []*url.URL{}
	for _, i := range foundURLs {
		found := false
		for _, j := range urlList {
			if i.String() == j.String() {
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
			if i.String() == j.String() {
				found = true
			}
		}
		// Dead upstream
		if !found {
			fmt.Println("Upstream died!", i.String())
			deadURLs = append(deadURLs, i)
		}
	}

	// Add new
	for _, i := range newURLs {
		rrb.AddTarget(&middleware.ProxyTarget{
			URL:  i,
			Name: i.String(),
		})
	}
	// Remove dead
	for _, i := range deadURLs {
		rrb.RemoveTarget(i.String())
	}

	// Copy over
	urlList = foundURLs
	fmt.Println("Using list:")
	for _, i := range urlList {
		fmt.Println(i.String())
	}
}

func StartServer() {
	Server = echo.New()
	Server.HideBanner = true
	Server.Use(middleware.Logger())

	rrb = middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{}) // No targets
	ScanForUpstreams()
	StartScanTicker()
	Server.GET("/upstream", ListUpstreams)
	Server.Group("/", middleware.Proxy(rrb)) // Hack to get / route working for proxy along with above route

	fmt.Println("Starting to serve proxy")
	Server.Start(":80")
}

func StartScanTicker() {
	fmt.Println("Starting scan ticker")
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

func ListUpstreams(c echo.Context) error {
	listString := ""
	for _, i := range urlList {
		listString += fmt.Sprintf("<li>%s</li>", i)
	}
	return c.HTML(200, fmt.Sprintf(`
		<h1>Upstreams</h1>
		<h2>Count: %d</h2>
		<h2>List:</h2>
		<ul>
		%s
		</ul>
	`, len(urlList), listString))
}
