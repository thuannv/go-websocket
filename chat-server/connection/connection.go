package connection

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	writeTimeout = 10 * time.Second
	pongTimeout  = 60 * time.Second
	pingPeriod   = pongTimeout * 9 / 10
	payloadSize  = 4096
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4024,
	WriteBufferSize: 4024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Connection is data structure holding information of a client connected via websocket
type Connection struct {
	userID         string
	ws             *websocket.Conn
	sendingChannel chan []byte
}

// ConnectionsManager is data structure for managing client websocket connections
type ConnectionsManager struct {
	connections map[string]*Connection
}

// Message struct for marshaling
type Message struct {
	ChannelID  string `json:"channel_id,omitempty"`
	SenderID   string `json:"sender_id,required"`
	ReceiverID string `json:"receiver_id,required"`
	Message    string `json:"Message,required"`
}

var connectionsManager = ConnectionsManager{make(map[string]*Connection)}

// Add new connection
func (cm *ConnectionsManager) Add(conn *Connection) {
	if conn != nil {
		cm.connections[conn.userID] = conn
	}
}

// Remove a connection
func (cm *ConnectionsManager) Remove(conn *Connection) {
	if conn != nil {
		delete(cm.connections, conn.userID)
	}
}

// Get connection associates with clientID
func (cm *ConnectionsManager) Get(clientID string) *Connection {
	conn, exist := cm.connections[clientID]
	if exist {
		return conn
	}
	return nil
}

func cleanUp(c *Connection) {
	if c != nil {
		connectionsManager.Remove(c)
		if c.ws != nil {
			c.ws.Close()
		}
		fmt.Println("Client id", c.userID, " is DISCONNECTED")
	}
}

func handleReceivingMessage(msg Message) {
	con := connectionsManager.Get(msg.ReceiverID)
	if con == nil {
		fmt.Println("Receiver ID:", msg.ReceiverID, "is OFFLINE.")
		return
	}

	fmt.Println("Receiver ID:", msg.ReceiverID, "is ONLINE -> sending message...")
	bytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("json.Marshal() failed, error:", err)
		return
	}
	con.sendingChannel <- bytes
}

func (c *Connection) readLoop() {
	defer cleanUp(c)
	c.ws.SetReadLimit(int64(payloadSize))
	c.ws.SetReadDeadline(time.Now().Add(pongTimeout))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongTimeout)); return nil })
	for {
		var msg Message
		err := c.ws.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v", err)
			}
			break
		}
		handleReceivingMessage(msg)
	}
}

func (c *Connection) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		cleanUp(c)
	}()
	for {
		select {
		case bytes, ok := <-c.sendingChannel:
			c.ws.SetWriteDeadline(time.Now().Add(writeTimeout))
			if !ok {
				fmt.Println("receiving NOT ok from sendingChannel -> sending CloseMessage and close connection.")
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, err := c.ws.NextWriter(websocket.BinaryMessage)
			if err != nil {
				fmt.Println("Failed to acquire BinaryMessage writer -> close connection.")
				return
			}
			fmt.Println("Write data to connection")
			writer.Write(bytes)
		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Function to handle incoming web socket connections
func HandleWebSocketConnections(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	uid, present := params["uid"]
	if !present {
		fmt.Println("missing uid")
		http.Error(w, "Missing uid", http.StatusBadRequest)
		return
	}
	clientID := uid[0]
	fmt.Println("Client id", clientID, "is CONNECTED")

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Fatal error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	connection := &Connection{userID: clientID, ws: ws, sendingChannel: make(chan []byte, payloadSize)}
	connectionsManager.Add(connection)
	go connection.readLoop()
	go connection.writeLoop()
}
