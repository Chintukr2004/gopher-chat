package main

import (
	"log"
	"net/http"
)

// / serveWs handles the initial HTTP request and upgrades it to a WebSocket

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	//upgrade the HTTP connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrador error:", err)
		return
	}

	//2. Create a new Client instance for this connection
	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256), // buffer of 256 messages

	}
	// /3. Register the new client with the hub
	client.hub.register <- client

	//4. Start the Read and Write pumps in seperate GOroutines!
	go client.writePump()
	go client.readPump()
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "index.html")
}

func main() {
	hub := NewHub()

	go hub.Run()

	http.HandleFunc("/", serveHome)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	log.Println("Server starting on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}
