package main

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type Hub struct {
	clients     map[*Client]bool
	boadcast    chan []byte
	register    chan *Client
	unregister  chan *Client
	redisClient *redis.Client
}

func NewHub() *Hub {

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	return &Hub{
		boadcast:    make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		redisClient: rdb,
	}
}

// his function listens to Redis for messages coming from ANY server instance
func (h *Hub) listenToRedis() {
	//subscribe to a central channel named "chat_channl"
	pubsub := h.redisClient.Subscribe(ctx, "chat_channel")
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		messageBytes := []byte(msg.Payload)

		for client := range h.clients {
			select {
			case client.send <- messageBytes:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}

	}

}

func (h *Hub) Run() {

	go h.listenToRedis()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.boadcast:
			// Save the message to Postgres first!
			SaveMessage(message)
			err := h.redisClient.Publish(ctx, "chat_channel", message).Err()
			if err != nil {
				log.Println("Error publishing to Redis:", err)
			}
		}
	}
}
