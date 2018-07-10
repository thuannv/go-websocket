package main

import (
	"fmt"
	"net/http"

	conf "./configure"
	connection "./connection"
)

func main() {
	c, ok := conf.LoadConfigs("configs", "conf")
	if !ok {
		fmt.Println("LoadConfigs name=configs, path=config failed.")
		return
	}

	fmt.Println("Load config success; configs:", c.String())

	address := fmt.Sprintf("%s:%d", c.WebSocketHost, c.WebSocketPort)

	http.HandleFunc("/ws", connection.HandleWebSocketConnections)
	fmt.Println("Server started listening at port", c.WebSocketPort)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Println("Start server failed.")
	}
}
