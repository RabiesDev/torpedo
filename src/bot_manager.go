package src

import (
	"fmt"
	"github.com/RabiesDev/go-logger"
	"github.com/shettyh/threadpool"
	"math/rand"
	"os"
	"time"
	"torpedo/src/utils"
)

type BotManager struct {
	Executor      *threadpool.ThreadPool
	Logger        *log.Logger
	ServerAddress string
	Proxies       []string
	Bots          []BotSnake
}

func (manager *BotManager) Prepare() {
	if manager.Logger == nil {
		manager.Logger = log.NewLogger(os.Stdout, log.DefaultPrefixes(), 2).WithColor()
	}

	manager.Executor = threadpool.NewThreadPool(600, 50000)
	manager.Bots = make([]BotSnake, 0)

	if len(manager.Proxies) > 0 {
		for _, proxy := range manager.Proxies {
			for i := 0; i < 2; i++ {
				manager.Bots = append(manager.Bots, BotSnake{
					Username: fmt.Sprintf("TWERKER_%v", rand.Intn(9999)),
					Proxy:    proxy,
				})
			}
		}
	} else {
		for i := 0; i < 4; i++ {
			manager.Bots = append(manager.Bots, BotSnake{
				Username: fmt.Sprintf("TWERKER_%v", rand.Intn(9999)),
			})
		}
	}
}

func (manager *BotManager) Execute() error {
	for i := range manager.Bots {
		bot := &manager.Bots[i]
		manager.Logger.Debugf("Connecting %s", bot.Username)
		bot.ConnectToServer(manager.ServerAddress, manager.Logger, manager.Executor)
		time.Sleep(time.Millisecond * 100)
	}

	for manager.IsAlive() {
		utils.FlushConsole()
		for i := range manager.Bots {
			bot := &manager.Bots[i]
			if bot.FailedAttempts <= 3 {
				if bot.Connected && !bot.Dead {
					bot.AngleTo(20000, 20000)
					manager.Logger.Infof("%s [X: %v, Y: %v, Angle: %v]",
						bot.Username,
						bot.PosX,
						bot.PosY,
						bot.Angle,
					)
				} else if !bot.Connecting {
					manager.Logger.Infof("%s trying to reconnect", bot.Username)
					bot.ConnectToServer(manager.ServerAddress, manager.Logger, manager.Executor)
				}
			} else {
				manager.Logger.Warnf("%s is currently in sleep mode", bot.Username)
			}
		}
		time.Sleep(time.Millisecond)
	}
	return nil
}

func (manager *BotManager) Destroy() {
	manager.Executor.Close()
}

func (manager *BotManager) IsAlive() bool {
	alive := false
	for i := range manager.Bots {
		bot := &manager.Bots[i]
		if bot.FailedAttempts <= 3 && (bot.Connecting || !bot.Dead) {
			alive = true
		}
	}
	return alive
}
