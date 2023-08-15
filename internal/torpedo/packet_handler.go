package torpedo

import (
	"errors"
	"fmt"
	"github.com/RabiesDev/go-logger"
	"math"
	"math/rand"
	"sync"
	"time"
	"torpedo/internal/helpers"
)

type PacketHandler struct {
	Logger   *log.Logger
	Mutex    sync.Mutex
	Handlers map[int]func(earthworm *Earthworm, data []int) [][]byte
}

func NewPacketHandler() *PacketHandler {
	logger := log.Default()
	logger.SetName("Handler")

	return &PacketHandler{
		Logger:   logger,
		Handlers: make(map[int]func(earthworm *Earthworm, data []int) [][]byte),
	}
}

func (packetHandler *PacketHandler) RegisterHandlers() {
	packetHandler.RegisterHandler([]int{54}, func(earthworm *Earthworm, data []int) [][]byte {
		initializePayload := make([]byte, 4+len(earthworm.Nickname))
		initializePayload[0] = 115
		initializePayload[1] = 10
		initializePayload[2] = byte(rand.Intn(42))
		initializePayload[3] = byte(len(earthworm.Nickname))
		for i, c := range earthworm.Nickname {
			initializePayload[4+i] = byte(c)
		}

		return [][]byte{
			helpers.DecodeSecret(data),
			initializePayload,
		}
	})
	packetHandler.RegisterHandler([]int{97}, func(earthworm *Earthworm, data []int) [][]byte {
		earthworm.Dead = false
		earthworm.NeedPing = true
		return nil
	})
	packetHandler.RegisterHandler([]int{112}, func(earthworm *Earthworm, data []int) [][]byte {
		earthworm.NeedPing = true
		return nil
	})
	packetHandler.RegisterHandler([]int{118}, func(earthworm *Earthworm, data []int) [][]byte {
		earthworm.Dead = true
		return nil
	})
	packetHandler.RegisterHandler([]int{115}, func(earthworm *Earthworm, data []int) [][]byte {
		if earthworm.UniqueId == 0 {
			earthworm.UniqueId = data[3]<<8 | data[4]
		}

		if earthworm.UniqueId == data[3]<<8|data[4] {
			if len(data) >= 31 {
				earthworm.Speed = int64((data[12]<<8 | data[13]) / 1e3)
				if ((((data[18] << 16) | (data[19] << 8) | data[20]) / 5) > 99) || ((((data[21] << 16) | (data[22] << 8) | data[23]) / 5) > 99) {
					earthworm.PosX = ((data[18] << 16) | (data[19] << 8) | data[20]) / 5
					earthworm.PosY = ((data[21] << 16) | (data[22] << 8) | data[23]) / 5
				}
			}
		}
		return nil
	})
	packetHandler.RegisterHandler([]int{110, 103}, func(earthworm *Earthworm, data []int) [][]byte {
		if earthworm.UniqueId == data[3]<<8|data[4] {
			earthworm.PosX = data[5]<<8 | data[6]
			earthworm.PosY = data[7]<<8 | data[8]

			if earthworm.LastPacket != nil {
				distance := int(earthworm.Speed * (time.Since(time.Now()) - *earthworm.LastPacket).Milliseconds() / 4)
				earthworm.PosX += int(math.Cos(earthworm.Angle)) * distance
				earthworm.PosY += int(math.Sin(earthworm.Angle)) * distance
			}

			earthworm.UpdateLastPacket()
		}
		return nil
	})
	packetHandler.RegisterHandler([]int{71, 78}, func(earthworm *Earthworm, data []int) [][]byte {
		if earthworm.UniqueId == data[3]<<8|data[4] {
			earthworm.PosX += data[5] - 128
			earthworm.PosY += data[6] - 128

			if earthworm.LastPacket != nil {
				distance := int(earthworm.Speed * (time.Since(time.Now()) - *earthworm.LastPacket).Milliseconds() / 4)
				earthworm.PosX += int(math.Cos(earthworm.Angle)) * distance
				earthworm.PosY += int(math.Sin(earthworm.Angle)) * distance
			}

			earthworm.UpdateLastPacket()
		}
		return nil
	})
}

func (packetHandler *PacketHandler) RegisterHandler(packetIds []int, handler func(earthworm *Earthworm, data []int) [][]byte) {
	defer packetHandler.Mutex.Unlock()
	packetHandler.Mutex.Lock()
	for _, packetId := range packetIds {
		packetHandler.Handlers[packetId] = handler
	}
}

func (packetHandler *PacketHandler) HandlePacket(earthworm *Earthworm, pid int, data []int) ([][]byte, error) {
	defer packetHandler.Mutex.Unlock()
	packetHandler.Mutex.Lock()

	handler, ok := packetHandler.Handlers[pid]
	if !ok {
		return nil, errors.New(fmt.Sprintf("No handler registered for PID %d", pid))
	}
	return handler(earthworm, data), nil
}
