package main

import (
	"context"
	"log"
	"net/http"
	"path"

	"github.com/gjae/wsmousetracker/server"
	"github.com/gjae/wsmousetracker/ws"
)

func main() {
	ctx, _ := context.WithCancel(context.Background())
	muxServer := server.NewServer("127.0.0.1:8090", nil)
	socket := ws.NewWebsocketHandler(muxServer, &ctx)

	muxServer.HandlerFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Dir("./templates/index.html"))
	})

	muxServer.HandlerFunc("/ws/", socket.UpgradeConnection(ctx))

	muxServer.Run(func() {
		log.Println("Server running shutdown routine")
	})
}
