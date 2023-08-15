package torpedo

import (
	"fmt"
	"github.com/RabiesDev/go-logger"
	"github.com/fasthttp/websocket"
	"math"
	"net/http"
	"net/url"
	"time"
	"torpedo/internal/helpers"
)

type Earthworm struct {
	Logger         *log.Logger
	Conn           *websocket.Conn
	RoutineTimer   *helpers.Stopwatch
	PingTimer      *helpers.Stopwatch
	LastPacket     *time.Duration
	RemoteProxy    string
	Nickname       string
	Connected      bool
	NeedPing       bool
	FailedAttempts int

	UniqueId  int
	Dead      bool
	Speed     int64
	PrevAngle float64
	Angle     float64
	PosX      int
	PosY      int
}

func NewEarthworm(remoteProxy string, nickname string) *Earthworm {
	logger := log.Default()
	logger.SetName(nickname)

	return &Earthworm{
		Logger:       logger,
		RoutineTimer: helpers.NewStopwatch(),
		PingTimer:    helpers.NewStopwatch(),
		RemoteProxy:  remoteProxy,
		Nickname:     nickname,
	}
}

func (earthworm *Earthworm) ConnectToServer(serverAddress string) error {
	earthworm.Logger.Debugln("Connecting to %s...", serverAddress)

	dialer := websocket.Dialer{
		Proxy: func(request *http.Request) (*url.URL, error) {
			if len(earthworm.RemoteProxy) > 0 {
				return url.Parse(earthworm.RemoteProxy)
			}
			return nil, nil
		},
		HandshakeTimeout:  time.Second * 10,
		EnableCompression: true,
	}

	conn, _, err := dialer.Dial(fmt.Sprintf("ws://%s/slither", serverAddress), http.Header{
		"origin":        {"http://slither.io"},
		"cache-control": {"no-cache"},
		"pragma":        {"no-cache"},
	})

	if err != nil {
		earthworm.Connected = false
		earthworm.FailedAttempts++
		return err
	}

	earthworm.Conn = conn
	earthworm.Connected = true
	earthworm.FailedAttempts = 0
	return nil
}

func (earthworm *Earthworm) ConnectionRoutine(packetHandler *PacketHandler) {
	defer func(Conn *websocket.Conn) {
		_ = Conn.Close()
	}(earthworm.Conn)
	earthworm.WritePacket([]byte{99})

	for earthworm.Conn != nil {
		messageType, payload, err := earthworm.Conn.ReadMessage()
		if err != nil {
			earthworm.Logger.Error(err.Error())
			earthworm.FailedAttempts++
			break
		} else if messageType != websocket.BinaryMessage {
			continue
		}

		data := make([]int, len(payload))
		for i, val := range payload {
			data[i] = int(val & 255)
		}

		payloads, err := packetHandler.HandlePacket(earthworm, data[2], data)
		if err != nil {
			packetHandler.Logger.Errorln(err.Error())
			continue
		} else if payloads == nil || len(payloads) == 0 {
			continue
		}

		for _, payload = range payloads {
			earthworm.WritePacket(payload)
		}
	}
}

func (earthworm *Earthworm) UpdateAndPing() {
	if earthworm.RoutineTimer.Finish(time.Millisecond*100) && earthworm.Angle != earthworm.PrevAngle {
		//earthworm.Logger.Debugln("%s updated angles to %v", earthworm.Nickname, math.Floor(earthworm.Angle))
		earthworm.WritePacket([]byte{byte(math.Floor(earthworm.Angle))})
		earthworm.PrevAngle = earthworm.Angle
		earthworm.RoutineTimer.Reset()
	}

	if earthworm.PingTimer.Finish(time.Millisecond*250) && earthworm.NeedPing {
		earthworm.WritePacket([]byte{251})
		earthworm.NeedPing = false
		earthworm.PingTimer.Reset()
	}
}

func (earthworm *Earthworm) AngleTo(x, y int) {
	diffX := float64(x - earthworm.PosX)
	diffY := float64(y - earthworm.PosY)
	newAngle := math.Atan(diffY / diffX)
	if diffX < 0 {
		newAngle += math.Pi
	}
	earthworm.Angle = math.Round((newAngle / (2 * math.Pi)) * 250)
}

func (earthworm *Earthworm) WritePacket(data []byte) {
	if err := earthworm.Conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		earthworm.Logger.Error(err.Error())
	}
}

func (earthworm *Earthworm) UpdateLastPacket() {
	now := time.Since(time.Now())
	earthworm.LastPacket = &now
}
