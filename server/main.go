package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Type string `json:"type"`
	Data string `json:"data"`
	To string `json:"to,omitempty"`
}

type Client struct {
	id string
	socket *websocket.Conn
	send chan []byte
	hub *Hub
}

type Hub struct {
	clients map[string]*Client
	register chan *Client
	unregister chan *Client
	broadcast chan []byte
}

var hub = Hub {
	register: make(chan *Client),
	unregister: make(chan *Client),
	clients: make(map[string]*Client),
	broadcast: make(chan []byte),
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.id] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client.id]; ok {
				delete(h.clients, client.id)
				close(client.send)
			}
		case message := <-h.broadcast:
			for id, client := range h.clients {
				select {
				case client.send <- message:
				default:
					log.Printf("Failed to send message to client %s", id)
					close(client.send)
					delete(h.clients, id)
				}
			} 
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{socket: socket, send: make(chan []byte, 256), hub: &hub}
	hub.register<- client
	go client.readPump()
	client.writePump()
}

func (c *Client) readPump() {
	defer func ()  {
		c.hub.unregister <- c
		c.socket.Close()
	}()
	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
                log.Println("WebSocket connection closed by client:", err)
            } else {
                log.Println("Error on message read:", err)
            }
			break
		}
		c.hub.broadcast<- message
	}
}

func (c *Client) writePump() {
	for {
		message, ok := <-c.send
		if !ok {
			c.socket.WriteMessage(websocket.CloseMessage, []byte{})
			break
		}
		c.socket.WriteMessage(websocket.TextMessage, message)
	}
} 

func main() {
	go hub.run()
	http.HandleFunc("/ws", handleWebSocket)
	log.Fatal(http.ListenAndServe(":8080", nil))
}