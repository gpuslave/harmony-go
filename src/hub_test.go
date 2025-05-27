package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings" // Added for TestServeWsClientNameAssignment
	"testing"
	"time"
	// "github.com/gorilla/websocket" // Not strictly needed for header setting if done manually
)

func TestNewHub(t *testing.T) {
	hub := newHub()
	if hub.broadcast == nil || hub.register == nil || hub.unregister == nil {
		t.Errorf("newHub channels not initialized")
	}
	if len(hub.clients) != 0 {
		t.Errorf("expected no clients, got %d", len(hub.clients))
	}
}

func TestHubRegisterBroadcastUnregister(t *testing.T) {
	hub := newHub()
	go hub.run()
	// Client names are assigned by serveWs in production, 
	// but for this test, we can assign them or leave them empty if the test logic doesn't depend on specific names.
	// The key is the message format.
	c1 := &Client{hub: hub, send: make(chan []byte, 1), Name: "client1"}
	c2 := &Client{hub: hub, send: make(chan []byte, 1), Name: "client2"}

	hub.register <- c1
	hub.register <- c2

	// Simulate a message already processed and prefixed by a client's readPump
	prefixedMsg := []byte("user123:hello")
	hub.broadcast <- prefixedMsg

	select {
	case got := <-c1.send:
		if !bytes.Equal(got, prefixedMsg) {
			t.Errorf("client1 got %s, want %s", got, prefixedMsg)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for client1 broadcast")
	}

	select {
	case got := <-c2.send:
		if !bytes.Equal(got, prefixedMsg) {
			t.Errorf("client2 got %s, want %s", got, prefixedMsg)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for client2 broadcast")
	}

	hub.unregister <- c1
	// After unregister, c1.send should be closed
	_, ok := <-c1.send
	if ok {
		t.Error("expected c1.send to be closed")
	}
}

func TestServeHome(t *testing.T) {
	// Successful GET request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	serveHome(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("serveHome GET /: code = %d, want %d", rr.Code, http.StatusOK)
	}
	if rr.Body.Len() == 0 {
		t.Error("serveHome GET /: empty body")
	}

	// Path not found
	req = httptest.NewRequest(http.MethodGet, "/foo", nil)
	rr = httptest.NewRecorder()
	serveHome(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("serveHome GET /foo: code = %d, want %d", rr.Code, http.StatusNotFound)
	}

	// Method not allowed
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	rr = httptest.NewRecorder()
	serveHome(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("serveHome POST /: code = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestServeWsClientNameAssignment(t *testing.T) {
	hub := newHub()
	go hub.run() // Hub needs to be running to process registrations

	// Make hub.register channel buffered to receive the client without blocking serveWs
	// This is a key change to allow capturing the client.
	hub.register = make(chan *Client, 1) // Buffer of 1

	// Create a mock request that would trigger serveWs
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	// Add WebSocket upgrade headers
	req.Header.Set("Connection", "upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "testkey") // A dummy key, required by upgrader

	rr := httptest.NewRecorder() // Mock response writer

	// Calling serveWs directly.
	// serveWs will attempt to upgrade, create a client, and register it.
	// The actual WebSocket connection isn't established beyond the upgrade handshake,
	// which is fine for testing client name assignment.
	serveWs(hub, rr, req)

	select {
	case client := <-hub.register:
		if client.Name == "" {
			t.Errorf("serveWs did not assign a name to the client")
		}
		if len(client.Name) != 7 {
			t.Errorf("client name should be 7 characters, got %d (name: '%s')", len(client.Name), client.Name)
		}
		// Check if the name contains only alphanumeric characters (optional, but good)
		for _, r := range client.Name {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
				t.Errorf("client name contains non-alphanumeric character: %s", client.Name)
				break
			}
		}
	case <-time.After(2 * time.Second): // Timeout if client isn't registered
		t.Errorf("timeout waiting for client to register in serveWs")
	}
}
