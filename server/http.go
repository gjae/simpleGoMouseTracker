package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

type OnStopServer func()

type ServerTimeoutConfig struct {
	WriteTimeout int64
	ReadTimeout  int64
	IdleTimeout  int64
}

type ServerHandler struct {
	router *mux.Router
	server *http.Server
}

func NewServerTimeoutConfig(wTimeout, rTimeout, idleTimeout int64) *ServerTimeoutConfig {
	return &ServerTimeoutConfig{
		WriteTimeout: wTimeout,
		ReadTimeout:  rTimeout,
		IdleTimeout:  idleTimeout,
	}
}

func NewServer(host string, config *ServerTimeoutConfig) *ServerHandler {
	handler := mux.NewRouter()

	if config == nil {
		config = NewServerTimeoutConfig(15, 15, 20)
	}

	return &ServerHandler{
		router: handler,
		server: &http.Server{
			Addr:         host,
			Handler:      handler,
			WriteTimeout: time.Duration(config.WriteTimeout) * time.Second,
			ReadTimeout:  time.Duration(config.ReadTimeout) * time.Second,
			IdleTimeout:  time.Duration(config.IdleTimeout) * time.Second,
		},
	}

}

func (srv *ServerHandler) Run(onStopServer OnStopServer) {
	go func(hs *http.Server) {
		if err := hs.ListenAndServe(); err != nil {
			log.Println("Server is stoped")
			if onStopServer != nil {
				onStopServer()
			}
		}
	}(srv.server)

	stopServerChan := make(chan os.Signal, 1)
	signal.Notify(stopServerChan, os.Interrupt)

	<-stopServerChan
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv.server.Shutdown(ctx)

	log.Println("Server is shutdown")
}

func (svr *ServerHandler) HandlerFunc(path string, handler func(w http.ResponseWriter, r *http.Request)) {
	svr.router.HandleFunc(path, handler)
}
