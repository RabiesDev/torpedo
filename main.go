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

	context := torpedo.NewWormManager(*serverAddress, proxies, 5*(len(proxies)+1))
	go torpedo.StartSyncServerLocation(context)

	context.RegisterWorms()
	context.StartConnect()
	go context.StartRoutine()

	time.Sleep(time.Second)
	context.WaitGroup.Wait()
}
