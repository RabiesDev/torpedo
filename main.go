package main

import (
	"flag"
	"time"
	"torpedo/internal/torpedo"
)

func main() {
	configPath := flag.String("config", "config.json", "")
	flag.Parse()

	config, err := torpedo.ParseConfig(*configPath)
	if err != nil {
		return
	}

	proxies := config.ParseProxies()
	context := torpedo.NewWormManager(config.ServerAddress, proxies, config.PointOffset, 5*(len(proxies)+1))
	go torpedo.StartSyncServerLocation(context)

	context.RegisterWorms()
	context.StartConnect()
	go context.StartRoutine()

	time.Sleep(time.Second)
	context.WaitGroup.Wait()
}
