package room

import (
	"log"

	"github.com/gorilla/websocket"
)

const sendBufferSize = 256

type Client struct {
	PlayerID string
	RoomCode string
	conn     *websocket.Conn
	send     chan []byte
	hub      *Hub
}

func NewClient(playerID, roomCode string, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		PlayerID: playerID,
		RoomCode: roomCode,
		conn:     conn,
		send:     make(chan []byte, sendBufferSize),
		hub:      hub,
	}
}

func (c *Client) Send(msg []byte) {
	select {
	case c.send <- msg:
	default:
		log.Printf("client %s send buffer full, dropping message", c.PlayerID)
	}
}

// ReadPump reads messages from the WebSocket and forwards them to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}
		c.hub.Incoming <- IncomingMessage{Client: c, Data: msg}
	}
}

// WritePump writes messages from the send channel to the WebSocket.
func (c *Client) WritePump() {
	defer c.conn.Close()

	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("write error for %s: %v", c.PlayerID, err)
			return
		}
	}
}
