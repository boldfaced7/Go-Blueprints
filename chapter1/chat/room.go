package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type room struct {
	// forward는 수신 메시지를 보관하는 채널이며
	// 수신한 메시지는 다른 클라이언트로 전달되어야
	forward chan []byte

	// join은 방에 들어오려는 클라이언트를 위한 채널
	join chan *client

	// leave는 방을 나가려는 클라이언트를 위한 채널
	leave chan *client

	// clients는 현재 채팅방에 있는 모든 클라이언트를 보유
	clients map[*client]bool
}

// newRoom은 새로운 방을 생성
func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// 입장
			r.clients[client] = true
		case client := <-r.leave:
			// 퇴장
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.forward:
			// 모든 클라이언트에게 메시지 전달
			for client := range r.clients {
				client.send <- msg
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize,
	WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
