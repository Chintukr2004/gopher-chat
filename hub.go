package main

type Hub struct {
	clients    map[*Client]bool
	boadcast   chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		boadcast:   make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
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
			for client := range h.clients {
				select {
				case client.send <- message:
				//message send succesfully
				default:
					// if client buffer is full, they stuck. Disconnect them to save server resources
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
