package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"text/template"

	"github.com/gjae/wsmousetracker/server"
	"github.com/gjae/wsmousetracker/ws"
)

//go:embed templates
var htmlTemplate embed.FS

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	muxServer := server.NewServer("127.0.0.1:8090", nil)
	socket := ws.NewWebsocketHandler(muxServer, &ctx)

	muxServer.HandlerFunc("/", func(w http.ResponseWriter, r *http.Request) {
		template, err := template.ParseFS(htmlTemplate, "templates/index.html")
		if err != nil {
			log.Println(err)
			return
		}
		_ = template.Execute(w, nil)
	})

	muxServer.HandlerFunc("/ws/", socket.UpgradeConnection(ctx))

	muxServer.Run(func() {
		log.Println("Server running shutdown routine")
		socket.Shutdown()
		cancel()
	})
}
