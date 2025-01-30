package main

import (
	// "bytes"
	// "log"
	// "net/http"
	// "time"

	"fmt"

	"github.com/gorilla/websocket"
)

const (
	// writeWait = 10 * time.Second

	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	hub *Hub

	conn *websocket.Conn

	send chan []byte
}

type Hub struct {
	clients map[*Client]bool

	broadcast chan []byte

	register chan *Client

	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func main() {
	fmt.Println("hello")
}
