package ws

import (
	"context"
	"log"
	"net/http"

	"github.com/gjae/wsmousetracker/server"
	"github.com/gorilla/websocket"
)

type WebsocketHandler struct {
	upgrade *websocket.Upgrader
	server  *server.ServerHandler
	hub     *Hub
}

func NewWebsocketHandler(server *server.ServerHandler, ctx *context.Context) *WebsocketHandler {

	handler := &WebsocketHandler{
		upgrade: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		server: server,
		hub:    NewClientHub(),
	}
	handler.hub.StartClientEvents(*ctx)
	return handler
}

func (ws *WebsocketHandler) UpgradeConnection(ctx context.Context) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Upgrading connection")
		c, err := ws.upgrade.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		go func(wsConn *websocket.Conn) {
			defer wsConn.Close()
			ws.hub.AddClient(wsConn)
			log.Println(r.URL.Query().Get("name"), r.URL.Query().Get("id"))
			ws.hub.SetUser(wsConn, r.URL.Query().Get("name"), r.URL.Query().Get("id"))
			ws.hub.BroadcastNewUser(wsConn)
			ws.hub.UpdateConnectedUsers(wsConn)
			for {
				var position Position
				err := wsConn.ReadJSON(&position)
				log.Println(position)
				if err != nil {
					log.Printf("User closed %v", err)
					ws.hub.RemoveClient(wsConn)
					return
				}
				ws.hub.SetUserPosition(wsConn, &position)
				ws.hub.PrintTrack(wsConn)
			}
		}(c)
	}
}

func (ws *WebsocketHandler) Shutdown() {

}
