package main

import (
	"flag"
	"time"
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

	wormManager := torpedo.NewWormManager(*serverAddress, proxies, 5*(len(proxies)+1))
	wormManager.RegisterWorms()
	wormManager.StartConnect()
	go wormManager.StartRoutine()

	time.Sleep(time.Second)
	wormManager.WaitGroup.Wait()
}
