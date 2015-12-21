package main

import (
	"zoop/proxyserver"
)


func main() {
	appserver := proxyserver.New()
	appserver.Init()
	appserver.Run()
}


