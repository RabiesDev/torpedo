package main

import (
	"flag"
	"torpedo/internal/helpers"
	"torpedo/internal/torpedo"
)

func main() {
	serverAddress := flag.String("server-address", "", "")
	proxiesPath := flag.String("proxies", "proxies.txt", "")
	flag.Parse()

	proxies, err := helpers.ReadLines(*proxiesPath)
	if err != nil {
		// TODO :: error handle
	}

	wormManager := torpedo.NewWormManager(*serverAddress, proxies, 10)
	go wormManager.ConnectWorms()
	go wormManager.StartRoutine()
	wormManager.WaitGroup.Wait()
}
