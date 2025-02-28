package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var upgrader = websocket.NewUpgrader()

func init() {
	upgrader.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		// echo
		c.WriteMessage(messageType, data)
	})
	upgrader.OnClose(func(c *websocket.Conn, err error) {
		log.Println("OnClose:", c.RemoteAddr().String(), err)
	})
}

func onHello(hrw http.ResponseWriter, req *http.Request) {
	hrw.WriteHeader(http.StatusOK)
	hrw.Write([]byte("hello world"))
}

func onWebsocket(hrw http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(hrw, req, nil)
	if err != nil {
		panic(err)
	}
	log.Println("OnOpen:", conn.RemoteAddr().String())
}

func main() {
	router := chi.NewRouter()

	router.Get("/hello", onHello)
	router.Get("/ws", onWebsocket)

	engine := nbhttp.NewEngine(nbhttp.Config{
		Network: "tcp",
		Addrs:   []string{"localhost:8080"},
	})
	engine.Handler = router

	err := engine.Start()
	if err != nil {
		log.Fatalf("nbio.Start failed: %v\n", err)
	}

	log.Println("serving [go-chi/chi] on [nbio]")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	engine.Shutdown(ctx)
}
