package ws

import (
	"context"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	ACTION_NEW_USER = "ACT_ADD_NEW_USER"
	REMOVE_USER     = "ACT_REMOVE_USER"
	UPDATE_POSITION = "ACT_UPDATE_POSITION"
)

type Position struct {
	X int64 `json:"X"`
	Y int64 `json:"Y"`
}

type Client struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	position *Position
	send     chan ResponseClient
	conn     *websocket.Conn
}

type Hub struct {
	clients    map[*websocket.Conn]*Client
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast  chan Position
	sync.Mutex
}

type ResponseClient struct {
	Position
	Client
	Action string `json:"action"`
}

func NewClientHub() *Hub {
	return &Hub{
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		clients:    make(map[*websocket.Conn]*Client),
		broadcast:  make(chan Position),
	}
}

func (c *Client) ClientBroadcast(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case message := <-c.send:
			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("Error writing message")
				return
			}

			log.Printf("Sent position")
		}
	}
}

func (hub *Hub) StartClientEvents(ctx context.Context) {
	go func(newClient <-chan *websocket.Conn) {

		for {
			select {
			case conn := <-newClient:
				hub.Lock()
				hub.clients[conn] = &Client{send: make(chan ResponseClient), conn: conn}
				log.Println("[NEW CLIENT]: ", conn.RemoteAddr())
				go hub.clients[conn].ClientBroadcast(ctx)
				hub.Unlock()
			case <-ctx.Done():
				log.Println("[NEW CLIENT] : CLOSE CHANNEL")
				close(hub.register)
				return
			}
		}

	}(hub.register)

	go func(unregister <-chan *websocket.Conn) {
		for {
			select {
			case deconn := <-unregister:
				hub.Lock()
				go hub.BroadCastRemoveUser(deconn)
				log.Println("[DECONN CLIENT]: ", deconn.RemoteAddr())
				hub.Unlock()
			case <-ctx.Done():
				log.Println("[DECONN CLIENT] : CLOSE CHANNEL")
				return
			}
		}

	}(hub.unregister)
}

func (hub *Hub) BroadCastRemoveUser(conn *websocket.Conn) {
	hub.Lock()
	defer hub.Unlock()
	for _, client := range hub.clients {
		client.send <- ResponseClient{Position: Position{X: 0, Y: 0}, Client: *hub.clients[conn], Action: REMOVE_USER}
	}

	conn.Close()
	delete(hub.clients, conn)
}

func (hub *Hub) AddClient(conn *websocket.Conn) {
	hub.register <- conn
}

func (hub *Hub) RemoveClient(conn *websocket.Conn) {
	hub.unregister <- conn
}

func (hub *Hub) Broadcast(message Position, conn *websocket.Conn) {
	hub.Lock()
	defer hub.Unlock()
	for _, client := range hub.clients {
		client.send <- ResponseClient{Position: message, Client: *hub.clients[conn], Action: UPDATE_POSITION}
	}
}

func (hub *Hub) BroadcastNewUser(conn *websocket.Conn) {
	hub.Lock()
	defer hub.Unlock()

	for _, client := range hub.clients {
		client.send <- ResponseClient{Position: Position{X: 43, Y: 137}, Client: *hub.clients[conn], Action: ACTION_NEW_USER}
	}
}

func (hub *Hub) SetUser(conn *websocket.Conn, name string, id string) {
	hub.Lock()
	defer hub.Unlock()
	hub.clients[conn].Name = name
	hub.clients[conn].ID = id

	log.Println("User setted: ", hub.clients[conn].Name, hub.clients[conn].ID)
}

func (hub *Hub) SetUserPosition(conn *websocket.Conn, position *Position) {
	hub.Lock()
	hub.clients[conn].position = position
	go hub.Broadcast(*position, conn)
	hub.Unlock()

}

func (hub *Hub) UpdateConnectedUsers(conn *websocket.Conn) {
	for _, client := range hub.clients {
		hub.clients[conn].send <- ResponseClient{Position: Position{X: 43, Y: 137}, Client: *client, Action: ACTION_NEW_USER}
	}
}

func (hub *Hub) PrintTrack(conn *websocket.Conn) {
	// log.Printf("Position.X: %d, Position.Y: %d", hub.clients[conn].position.X, hub.clients[conn].position.Y)
}

func (hub *Hub) CleanupConnections() {
	hub.Lock()
	defer hub.Unlock()
	for _, client := range hub.clients {
		hub.unregister <- client.conn
	}

}
