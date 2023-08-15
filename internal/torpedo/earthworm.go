package torpedo

import (
	"fmt"
	"github.com/RabiesDev/go-logger"
	"github.com/fasthttp/websocket"
	"math"
	"net/http"
	"net/url"
	"strconv"
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
	Status         string
	Connected      bool
	Initialized    bool
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
	return &Earthworm{
		Logger:       log.Default().WithColor(),
		RoutineTimer: helpers.NewStopwatch(),
		PingTimer:    helpers.NewStopwatch(),
		RemoteProxy:  remoteProxy,
		Nickname:     nickname,
	}
}

func (earthworm *Earthworm) ConnectToServer(serverAddress string) error {
	earthworm.Logger.Debugln(fmt.Sprintf("Connecting to %s...", serverAddress))
	earthworm.Status = "Connecting"

	dialer := websocket.Dialer{
		Proxy: func(request *http.Request) (*url.URL, error) {
			if len(earthworm.RemoteProxy) > 0 {
				return url.Parse(earthworm.RemoteProxy)
			}
			return http.ProxyFromEnvironment(request)
		},
		HandshakeTimeout:  time.Second * 10,
		EnableCompression: true,
	}

	conn, _, err := dialer.Dial(fmt.Sprintf("ws://%s/slither", serverAddress), http.Header{
		"User-Agent":      {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"},
		"Accept-Encoding": {"gzip, deflate"},
		"origin":          {"http://slither.io"},
		"cache-control":   {"no-cache"},
		"pragma":          {"no-cache"},
	})

	if err != nil {
		earthworm.Status = fmt.Sprintf("Disconnected (%s)", err.Error())
		earthworm.Connected = false
		earthworm.FailedAttempts++
		return err
	}

	earthworm.Conn = conn
	earthworm.Status = "Connected"
	earthworm.Connected = true
	earthworm.FailedAttempts = 0
	return nil
}

func (earthworm *Earthworm) ConnectionRoutine(packetHandler *PacketHandler) {
	defer func(Conn *websocket.Conn) {
		earthworm.Logger.Debugln("Connection closed")
		_ = Conn.Close()
	}(earthworm.Conn)

	// hello packet
	earthworm.WritePacket([]byte{99})

	for earthworm.Conn != nil && earthworm.Connected {
		messageType, payload, err := earthworm.Conn.ReadMessage()
		if err != nil {
			earthworm.Status = fmt.Sprintf("Disconnected (%s)", err.Error())
			earthworm.Connected = false
			earthworm.Initialized = false
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
	if !earthworm.Connected || !earthworm.Initialized || earthworm.Dead {
		return
	}

	if earthworm.RoutineTimer.Finish(time.Millisecond*100) && earthworm.Angle != earthworm.PrevAngle {
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
	if earthworm.Conn == nil {
		return
	}

	if err := earthworm.Conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		earthworm.Logger.Error(err.Error())
	}
}

func (earthworm *Earthworm) UpdateLastPacket() {
	now := time.Since(time.Now())
	earthworm.LastPacket = &now
}

func (earthworm *Earthworm) ToString() string {
	return fmt.Sprintf("%s [Status: %s, X: %s, Y: %s, Angle: %s]",
		string(earthworm.Logger.ApplyColor([]byte(earthworm.Nickname), log.Bold)),
		string(earthworm.Logger.ApplyColor([]byte(earthworm.Status), log.Cyan)),
		string(earthworm.Logger.ApplyColor([]byte(strconv.Itoa(earthworm.PosX)), log.Blue)),
		string(earthworm.Logger.ApplyColor([]byte(strconv.Itoa(earthworm.PosY)), log.Blue)),
		string(earthworm.Logger.ApplyColor([]byte(strconv.FormatFloat(earthworm.Angle, 'f', -1, 64)), log.Orange)),
	)
}
