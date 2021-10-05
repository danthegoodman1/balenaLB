package main

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
)

var (
	Server     *echo.Echo
	urlList    []url.URL
	cidrPrefix = "192.168.1"
)

func ScanForUpstreams() {
	for i := 0; i < 255; i++ {
		fmt.Println("checking", fmt.Sprintf("%s.%d", cidrPrefix, i))
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(fmt.Sprintf("%s.%d", cidrPrefix, i), "80"), time.Millisecond*200)
		if err != nil {
			fmt.Println("Failed to connect to", net.JoinHostPort(fmt.Sprintf("%s.%d", cidrPrefix, i), "80"))
			fmt.Println(err)
		}
		if conn != nil {
			fmt.Println("Connected to device", fmt.Sprintf("%s.%d", cidrPrefix, i))
			conn.Close()
		}
		time.Sleep(time.Millisecond * 100)
	}
	time.Sleep(time.Second * 10)
}

func StartServer() {
	Server = echo.New()
	Server.HideBanner = true

	Server.Start(":8080")
}
