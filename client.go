package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	//time allowed to write a message to peer.
	writeWait = 10 * time.Second
	// time allowed to read the nexr pong message from the peer.
	PongWait = 60 * time.Second
	// senf to peer with this period.
	PingPeriod = (PongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // allow all origins

}

type Client struct {
	hub  *Hub
	conn *websocket.Conn //websocket connection
	send chan []byte     // buffered channel of outbound messages
}

// this will pumps messages from the websocket connection to the hub

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error %v", err)
			}
			break
		}
		// send the message to the Hub's broadcast channel
		c.hub.boadcast <- message
	}

}

// this pumps message from the hub to the websocket connection

func (c *Client) writePump() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// the hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil{
				return 
			}
			w.Write(message)
			if err := w.Close(); err != nil{
				return
			}
		
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil{
				return 
			}

		}
	}

}
