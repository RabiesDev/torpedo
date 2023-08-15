package torpedo

import (
	"fmt"
	"github.com/RabiesDev/go-logger"
	"math/rand"
	"strconv"
	"sync"
	"time"
	"torpedo/internal/helpers"
)

type EarthwormManager struct {
	Logger        *log.Logger
	WaitGroup     *sync.WaitGroup
	TimeoutTimer  *helpers.Stopwatch
	FlushTimer    *helpers.Stopwatch
	PacketHandler *PacketHandler
	EarthwormPool chan struct{}
	ConnectedChan chan *Earthworm
	Earthworms    []*Earthworm
	Proxies       []string
	ServerAddress string
	Limit         int
}

func NewWormManager(serverAddress string, proxies []string, limit int) *EarthwormManager {
	packetHandler := NewPacketHandler()
	packetHandler.RegisterHandlers()

	return &EarthwormManager{
		Logger:        log.Default().WithColor(),
		WaitGroup:     new(sync.WaitGroup),
		TimeoutTimer:  helpers.NewStopwatch(),
		FlushTimer:    helpers.NewStopwatch(),
		PacketHandler: packetHandler,
		EarthwormPool: make(chan struct{}, limit),
		ConnectedChan: make(chan *Earthworm, 100),
		Earthworms:    make([]*Earthworm, 0),
		Proxies:       proxies,
		ServerAddress: serverAddress,
		Limit:         limit,
	}
}

func (context *EarthwormManager) RegisterWorms() {
	registerWorms := func(context *EarthwormManager, proxy string) {
		for i := 0; i < 2; i++ {
			nickname := fmt.Sprintf("Torpedo_%v", rand.Intn(9999))
			context.Earthworms = append(context.Earthworms, NewEarthworm(proxy, nickname))
		}
	}

	if len(context.Proxies) == 0 {
		registerWorms(context, "")
	} else {
		for _, proxy := range context.Proxies {
			registerWorms(context, proxy)
		}
	}
}

func (context *EarthwormManager) ConnectWorms() {
	for _, earthworm := range context.Earthworms {
		context.WaitGroup.Add(1)
		context.EarthwormPool <- struct{}{}

		go func(earthworm *Earthworm) {
			defer func() {
				<-context.EarthwormPool
				context.WaitGroup.Done()
			}()

			err := earthworm.ConnectToServer(context.ServerAddress)
			if err != nil {
				earthworm.Logger.Errorln(err.Error())
				return
			}

			context.ConnectedChan <- earthworm
			context.TimeoutTimer.Reset()
		}(earthworm)
	}
}

func (context *EarthwormManager) StartRoutine() {
	for {
		if context.FlushTimer.Finish(time.Millisecond * 200) {
			helpers.FlushConsole()
			context.FlushTimer.Reset()
		}

		select {
		case earthworm := <-context.ConnectedChan:
			if earthworm.Dead {
				break
			}

			context.WaitGroup.Add(1)
			go func(earthworm *Earthworm) {
				defer context.WaitGroup.Done()
				go earthworm.ConnectionRoutine(context.PacketHandler)
				for !earthworm.Dead && earthworm.Connected {
					context.Logger.Infoln(fmt.Sprintf("%s [Status: %s, X: %s, Y: %s, Angle: %s]",
						string(context.Logger.ApplyColor([]byte(earthworm.Nickname), log.Bold)),
						string(context.Logger.ApplyColor([]byte(earthworm.Status), log.Cyan)),
						string(context.Logger.ApplyColor([]byte(strconv.Itoa(earthworm.PosX)), log.Blue)),
						string(context.Logger.ApplyColor([]byte(strconv.Itoa(earthworm.PosY)), log.Blue)),
						string(context.Logger.ApplyColor([]byte(strconv.FormatFloat(earthworm.Angle, 'f', -1, 64)), log.Orange))),
					)

					context.TimeoutTimer.Reset()
					earthworm.UpdateAndPing()
					earthworm.AngleTo(20000, 20000)
					time.Sleep(time.Millisecond * 200)
				}
			}(earthworm)
		default:
			if context.TimeoutTimer.Finish(time.Second * 10) {
				return // Terminate StartRoutine when all worms are dead
			}
		}
	}
}
