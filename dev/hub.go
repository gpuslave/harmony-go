package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"

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

func (hub *Hub) run() {
	for {
		select {
		case client := <-hub.register:
			hub.clients[client] = true
		case client := <-hub.unregister:
			if _, ok := hub.clients[client]; ok {
				delete(hub.clients, client)
				close(client.send)
			}
		case message := <-hub.broadcast:
			for client := range hub.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(hub.clients, client)
				}
			}
		}

	}
}

var addr = flag.String("addr", ":8080", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	testFilePath := filepath.Join("web", "html", "test.html")
	// homeFilePath := filepath.Join("web", "html", "home.html")

	http.ServeFile(w, r, testFilePath)
}

func main() {
	fmt.Println("Init server...")
	flag.Parse()
	hub := newHub()
	go hub.run()
	fmt.Println("Hub successfully run.")

	http.HandleFunc("/", serveHome)

	// http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
	// 	serveWs(hub, w, r)
	// })

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("Listend and serve", err)
	}

	fmt.Println("done.")
}
