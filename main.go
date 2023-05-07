package main

import (
	"bufio"
	"os"
	torpedo "torpedo/src"
)

func main() {
	serverAddress := "209.58.183.136:444"
	if len(os.Args) > 1 {
		serverAddress = os.Args[len(os.Args)-1]
	}

	proxies := make([]string, 0)
	if file, err := os.Open("proxies.txt"); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			proxies = append(proxies, scanner.Text())
		}
	}

	context := torpedo.BotManager{
		ServerAddress: serverAddress,
		Proxies:       proxies,
	}
	context.Prepare()
	if err := context.Execute(); err != nil {
		panic(err)
	}
}
