package torpedo

import (
	"encoding/json"
	"github.com/fasthttp/websocket"
	"net/http"
	"strconv"
	"time"
	"torpedo/internal/helpers"
)

type SyncPayload struct {
	Status string `json:"status"`
	PointX string `json:"pointX"`
	PointY string `json:"pointY"`
}

func StartSyncServerLocation(context *EarthwormManager) {
	updateTimer := helpers.NewStopwatch()
	disconnected := false

	dialer := websocket.Dialer{
		HandshakeTimeout: time.Second * 10,
	}

	conn, _, err := dialer.Dial("ws://127.0.0.1:1337", http.Header{})
	if err != nil {
		context.Logger.Errorln("Location synchronization failed")
		return
	}

	defer func(Conn *websocket.Conn) {
		context.Logger.Infoln("Finish location synchronization")
		_ = Conn.Close()
		disconnected = true
	}(conn)

	go func(Conn *websocket.Conn) {
		for conn != nil && !disconnected {
			if updateTimer.Finish(time.Millisecond * 200) {
				_ = conn.WriteMessage(websocket.TextMessage, []byte{})
				updateTimer.Reset()
			}
		}
	}(conn)

	for conn != nil && !disconnected {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			context.Logger.Error(err.Error())
			return
		} else if len(payload) == 0 {
			continue
		}

		var syncPayload SyncPayload
		if err = json.Unmarshal(payload, &syncPayload); err != nil {
			continue
		}

		switch syncPayload.Status {
		case "established":
			if len(syncPayload.PointX) == 0 || len(syncPayload.PointY) == 0 {
				_ = conn.WriteMessage(websocket.TextMessage, []byte{})
				continue
			}

			pointX, _ := strconv.Atoi(syncPayload.PointX)
			pointY, _ := strconv.Atoi(syncPayload.PointY)
			context.SetPoints(pointX, pointY)
			break

		case "dead":
			context.Logger.Debugln("Target has died")
			break
		}
	}
}
