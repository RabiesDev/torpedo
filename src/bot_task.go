package src

import (
	"fmt"
	"github.com/RabiesDev/go-logger"
	"github.com/fasthttp/websocket"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"time"
	"torpedo/src/utils"
)

type BotTask struct {
	ServerAddress string
	Logger        *log.Logger
	Bot           *BotSnake

	conn       *websocket.Conn
	lastPacket *time.Duration
	snakeSpeed int64
	needPing   bool
}

func (task *BotTask) UpdateLastPacket() {
	now := time.Since(time.Now())
	task.lastPacket = &now
}

func (task *BotTask) WritePacket(buffer []byte) {
	if err := task.conn.WriteMessage(websocket.BinaryMessage, buffer); err != nil {
		task.Logger.Error(err.Error())
	}
}

func (task *BotTask) Run() {
	dialer := websocket.Dialer{
		Proxy: func(request *http.Request) (*url.URL, error) {
			if len(task.Bot.Proxy) > 0 {
				return &url.URL{
					Host:   task.Bot.Proxy,
					Scheme: "socks5",
				}, nil
			}
			return nil, nil
		},
		HandshakeTimeout:  10 * time.Second,
		EnableCompression: true,
	}
	conn, _, err := dialer.Dial(fmt.Sprintf("ws://%s/slither", task.ServerAddress), http.Header{
		"origin":        {"http://slither.io"},
		"cache-control": {"no-cache"},
		"pragma":        {"no-cache"},
	})
	if err != nil {
		//task.Logger.Error(err.Error())
		task.Bot.Connecting = false
		task.Bot.Connected = false
		task.Bot.FailedAttempts++
		return
	}

	task.conn = conn
	defer func(conn *websocket.Conn) {
		// not throwing an error
		task.Logger.Debugf("Connection %s Closed", task.Bot.Username)
		task.Bot.Connecting = false
		task.Bot.Connected = false
		_ = conn.Close()
	}(conn)

	task.Bot.Connecting = false
	task.Bot.Connected = true
	task.Logger.Infof("Connected %s", task.Bot.Username)
	task.WritePacket([]byte{99})

	updaterStopwatch := utils.NewStopwatch()
	pingerStopwatch := utils.NewStopwatch()
	go func() {
		for {
			if task.Bot.Connected {
				if task.Bot.Dead {
					return
				}

				if updaterStopwatch.Finish(time.Millisecond*100) && task.Bot.Angle != task.Bot.PrevAngle {
					task.Logger.Debugf("%s updated angles to %v", task.Bot.Username, math.Floor(task.Bot.Angle))
					task.WritePacket([]byte{byte(math.Floor(task.Bot.Angle))})
					task.Bot.PrevAngle = task.Bot.Angle
					updaterStopwatch.Reset()
				}

				if pingerStopwatch.Finish(time.Millisecond*250) && task.needPing {
					task.WritePacket([]byte{251})
					task.needPing = false
					pingerStopwatch.Reset()
				}
			}
			time.Sleep(time.Millisecond)
		}
	}()

	for conn != nil {
		_, bytes, err := conn.ReadMessage()
		if err != nil {
			task.Logger.Error(err.Error())
			task.Bot.FailedAttempts++
			return
		}

		buffer := make([]int, len(bytes))
		for i := 0; i < len(bytes); i++ {
			buffer[i] = int(bytes[i] & 255)
		}

		switch buffer[2] {
		case 54: // prepare
			setupPayload := make([]byte, 4+len(task.Bot.Username))
			setupPayload[0] = 115
			setupPayload[1] = 11 - 1
			setupPayload[2] = byte(rand.Intn(42))
			setupPayload[3] = byte(len(task.Bot.Username))
			for i, c := range task.Bot.Username {
				setupPayload[4+i] = byte(c)
			}

			task.WritePacket(utils.DecodeSecret(buffer))
			task.WritePacket(setupPayload)

		case 97: // init
			task.Bot.Dead = false
			task.needPing = true

		case 112: // pong
			task.needPing = true

		case 115: // add/remove snake
			if task.Bot.UniqueId == 0 {
				task.Bot.UniqueId = buffer[3]<<8 | buffer[4]
			}
			if task.Bot.UniqueId == buffer[3]<<8|buffer[4] {
				if len(buffer) >= 31 {
					task.snakeSpeed = int64((buffer[12]<<8 | buffer[13]) / 1e3)
					if ((((buffer[18] << 16) | (buffer[19] << 8) | buffer[20]) / 5) > 99) || ((((buffer[21] << 16) | (buffer[22] << 8) | buffer[23]) / 5) > 99) {
						task.Bot.PosX = ((buffer[18] << 16) | (buffer[19] << 8) | buffer[20]) / 5
						task.Bot.PosY = ((buffer[21] << 16) | (buffer[22] << 8) | buffer[23]) / 5
					}
				}
			}

		case 110, 103: // update pos
			if task.Bot.UniqueId == buffer[3]<<8|buffer[4] {
				task.Bot.PosX = buffer[5]<<8 | buffer[6]
				task.Bot.PosY = buffer[7]<<8 | buffer[8]

				if task.lastPacket != nil { // fixer
					distance := int(task.snakeSpeed * (time.Since(time.Now()) - *task.lastPacket).Milliseconds() / 4)
					task.Bot.PosX += int(math.Cos(task.Bot.Angle)) * distance
					task.Bot.PosY += int(math.Sin(task.Bot.Angle)) * distance
				}
				task.UpdateLastPacket()
			}

		case 71, 78: // update pos
			if task.Bot.UniqueId == buffer[3]<<8|buffer[4] {
				task.Bot.PosX += buffer[5] - 128
				task.Bot.PosY += buffer[6] - 128

				if task.lastPacket != nil { // fixer
					distance := int(task.snakeSpeed * (time.Since(time.Now()) - *task.lastPacket).Milliseconds() / 4)
					task.Bot.PosX += int(math.Cos(task.Bot.Angle)) * distance
					task.Bot.PosY += int(math.Sin(task.Bot.Angle)) * distance
				}
				task.UpdateLastPacket()
			}

		case 118: // dead
			task.Logger.Infof("%s has died", task.Bot.Username)
			task.Bot.Dead = true
			return

		default:
			// unknown packet
			//task.Logger.Infof("Received unknown pid: %v", buffer[2])
		}
	}
}
