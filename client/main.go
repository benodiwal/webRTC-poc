package main

import (
	"bufio"
	"log"
	"os"

	"github.com/gorilla/websocket"
)

func main() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket server:", err)
	}
	defer conn.Close()

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading message:", err)
				break
			}
			log.Println("Received message:", string(message))
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			continue
		}
		err = conn.WriteMessage(websocket.TextMessage, []byte(text))
		if err != nil {
			log.Println("Error sending message:", err)
			break
		}
	}
}
