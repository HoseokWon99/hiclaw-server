package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"hiclaw-server/core"

	"github.com/coder/websocket"
)

type WSConn struct {
	conn *websocket.Conn
}

func NewWSConn(conn *websocket.Conn) *WSConn {
	return &WSConn{conn: conn}
}

func (w *WSConn) Send(msg *core.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return w.conn.Write(context.Background(), websocket.MessageText, data)
}

func (w *WSConn) ReadMessage() (*core.Message, error) {
	_, data, err := w.conn.Read(context.Background())
	if err != nil {
		return nil, err
	}
	var msg core.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (w *WSConn) Close() error {
	return w.conn.CloseNow()
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("ws accept: %v", err)
		return
	}

	deviceIP := r.Header.Get("X-Device-IP")
	deviceName := r.Header.Get("X-Device-Name")
	if deviceIP == "" {
		deviceIP = r.RemoteAddr
	}

	wsConn := NewWSConn(conn)
	h.hub.Add(deviceIP, wsConn)
	defer func() {
		h.hub.Remove(deviceIP)
		wsConn.Close()
	}()

	log.Printf("ws connected: %s (%s)", deviceName, deviceIP)

	for {
		msg, err := wsConn.ReadMessage()
		if err != nil {
			log.Printf("ws disconnected: %s (%v)", deviceIP, err)
			return
		}
		h.chatSvc.SendMessage(
			&core.Device{IP: deviceIP, Name: deviceName},
			msg.Content,
		)
	}
}
