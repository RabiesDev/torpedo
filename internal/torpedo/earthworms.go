package torpedo

import (
	"github.com/RabiesDev/go-logger"
	"sync"
	"time"
	"torpedo/internal/helpers"
)

type EarthwormManager struct {
	Logger        *log.Logger
	WaitGroup     *sync.WaitGroup
	TimeoutTimer  *helpers.Stopwatch
	PacketHandler *PacketHandler
	EarthwormPool chan struct{}
	ConnectedChan chan *Earthworm
	Earthworms    []*Earthworm
	Proxies       []string
	ServerAddress string
	Limit         int
}

func NewWormManager(serverAddress string, proxies []string, limit int) *EarthwormManager {
	return &EarthwormManager{
		Logger:        log.Default(),
		TimeoutTimer:  helpers.NewStopwatch(),
		PacketHandler: NewPacketHandler(),
		EarthwormPool: make(chan struct{}, limit),
		ConnectedChan: make(chan *Earthworm, 100),
		Earthworms:    make([]*Earthworm, 0),
		Proxies:       proxies,
		ServerAddress: serverAddress,
		Limit:         limit,
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

			context.TimeoutTimer.Reset()
		}(earthworm)
	}
}

func (context *EarthwormManager) StartRoutine() {
	for {
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
					context.TimeoutTimer.Reset()
					earthworm.UpdateAndPing()
					earthworm.AngleTo(0, 0)
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
