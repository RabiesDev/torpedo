package src

import (
	"github.com/RabiesDev/go-logger"
	"github.com/shettyh/threadpool"
	"math"
)

type BotSnake struct {
	Proxy          string
	Connecting     bool
	Connected      bool
	FailedAttempts int

	UniqueId  int
	Username  string
	PrevAngle float64
	Angle     float64
	PosX      int
	PosY      int
	Dead      bool
}

func (bot *BotSnake) ConnectToServer(serverAddress string, logger *log.Logger, executor *threadpool.ThreadPool) {
	if bot.FailedAttempts > 3 {
		return
	}

	bot.Connecting = true
	bot.Connected = false
	if err := executor.Execute(&BotTask{
		ServerAddress: serverAddress,
		Logger:        logger,
		Bot:           bot,
	}); err != nil {
		bot.Connecting = false
		bot.FailedAttempts++
	}
}

func (bot *BotSnake) AngleTo(x, y int) {
	diffX := float64(x - bot.PosX)
	diffY := float64(y - bot.PosY)
	angle := math.Atan(diffY / diffX)
	if diffX < 0 {
		angle += math.Pi
	}
	bot.Angle = math.Round((angle / (2 * math.Pi)) * 250)
}
