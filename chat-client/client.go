package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// Message struct for marshaling
type Message struct {
	ChannelID  string `json:"channel_id,omitempty"`
	SenderID   string `json:"sender_id,required"`
	ReceiverID string `json:"receiver_id,required"`
	Message    string `json:"Message,required"`
}

func main() {
	addr := flag.String("addr", "localhost:8000", "http service address")
	senderID := flag.String("senderId", "", "SenderID to identify to server")
	receiverID := flag.String("receiverId", "", "ReceiverID to identify to server")
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	rawQuery := fmt.Sprintf("uid=%s", *senderID)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws", RawQuery: rawQuery}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			var message Message
			err := c.ReadJSON(&message)
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	sendingMessage := Message{SenderID: *senderID, ReceiverID: *receiverID, Message: "Hello there"}

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			writer, error := c.NextWriter(websocket.BinaryMessage)
			if error != nil {
				log.Println("Failed to acquire writer")
				return
			}
			data, err := json.Marshal(sendingMessage)
			if err != nil {
				log.Println("Failed to marshal json")
				return
			}
			writer.Write(data)

		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
