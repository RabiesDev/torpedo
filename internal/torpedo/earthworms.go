package torpedo

import (
	"fmt"
	"github.com/RabiesDev/go-logger"
	"math/rand"
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
	PointOffset   int
	Limit         int

	PointMutex sync.Mutex
	PointX     int
	PointY     int
}

func NewWormManager(serverAddress string, proxies []string, pointOffset, limit int) *EarthwormManager {
	packetHandler := NewPacketHandler()
	packetHandler.RegisterHandlers()

	return &EarthwormManager{
		Logger:        log.Default().WithColor(),
		WaitGroup:     new(sync.WaitGroup),
		TimeoutTimer:  helpers.NewStopwatch(),
		PacketHandler: packetHandler,
		EarthwormPool: make(chan struct{}, limit),
		ConnectedChan: make(chan *Earthworm, limit),
		Earthworms:    make([]*Earthworm, 0),
		Proxies:       proxies,
		ServerAddress: serverAddress,
		PointOffset:   pointOffset,
		Limit:         limit,
		PointX:        20000,
		PointY:        20000,
	}
}

func (context *EarthwormManager) RegisterWorms() {
	registerWorms := func(context *EarthwormManager, proxy string) {
		for i := 0; i < 3; i++ {
			nickname := fmt.Sprintf("Torpedo_%v", rand.Intn(99999))
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

func (context *EarthwormManager) StartConnect() {
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
		select {
		case earthworm := <-context.ConnectedChan:
			context.WaitGroup.Add(1)
			go func(earthworm *Earthworm) {
				defer context.WaitGroup.Done()
				go earthworm.ConnectionRoutine(context.PacketHandler)
				for !earthworm.Dead && earthworm.Connected {
					context.Logger.Infoln(earthworm.ToString())
					context.TimeoutTimer.Reset()
					context.PointMutex.Lock()
					earthworm.UpdateAndPing()
					earthworm.AngleTo(context.PointX, context.PointY)
					context.PointMutex.Unlock()
					time.Sleep(time.Millisecond * 100)
				}

				context.Logger.Infoln(earthworm.ToString())
				context.ReviveAndConnect(earthworm)
			}(earthworm)
		default:
			if context.TimeoutTimer.Finish(time.Second * 10) {
				return // Terminate StartRoutine when all worms are dead
			}
		}
	}
}

func (context *EarthwormManager) ReviveAndConnect(earthworm *Earthworm) {
	context.Logger.Infoln(fmt.Sprintf("Reviving %s and connecting...", earthworm.Nickname))

	earthworm.Initialized = false
	earthworm.Connected = false
	earthworm.Dead = false
	earthworm.UniqueId = 0
	earthworm.Conn = nil

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

func (context *EarthwormManager) SetPoints(x, y int) {
	defer context.PointMutex.Unlock()
	context.PointMutex.Lock()
	context.Logger.Debugln(fmt.Sprintf("Synchronization [X: %v, Y: %v]", x, y))
	context.PointX = x + context.PointOffset
	context.PointY = y + context.PointOffset
}
