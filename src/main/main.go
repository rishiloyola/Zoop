package main

import (
	"fmt"
	"proxyserver"
)

func main() {
	c := proxyserver.New()
	c.Init("/HTTPserver", "/pushpin", "127.0.0.1:2181")
	fmt.Println("started")
	c.Run()
}
