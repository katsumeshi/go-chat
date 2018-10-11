package main

import (
	"log"
	"net/http"

	"github.com/stretchr/objx"

	"../trace"

	"github.com/gorilla/websocket"
)

type room struct {
	forward chan *message
	join    chan *client
	leave   chan *client
	clients map[*client]bool
	tracer  trace.Tracer
	avatar  Avatar
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			r.tracer.Trace("New Client joined")
		case client := <-r.leave:
			delete(r.clients, client)
			r.tracer.Trace("Client left")
		case msg := <-r.forward:
			r.tracer.Trace("Message received: ", msg.Message)
			for client := range r.clients {
				log.Println("case msg := <-r.forward: ", msg)
				client.send <- msg
				r.tracer.Trace(" -- sent to client")
			}
		}
	}
}

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServerHttp:", err)
		return
	}
	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatal("Failed to get auth cookie:", err)
		return
	}
	client := &client{
		socket:   socket,
		send:     make(chan *message, messageBufferSize),
		room:     r,
		userData: objx.MustFromBase64(authCookie.Value),
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}

func newRoom(avatar Avatar) *room {
	return &room{
		forward: make(chan *message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
		avatar:  avatar,
	}
}
