package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
	c1 := &Client{hub: hub, send: make(chan []byte, 1)}
	c2 := &Client{hub: hub, send: make(chan []byte, 1)}

	hub.register <- c1
	hub.register <- c2

	msg := []byte("hello")
	hub.broadcast <- msg

	select {
	case got := <-c1.send:
		if !bytes.Equal(got, msg) {
			t.Errorf("client1 got %s, want %s", got, msg)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for client1 broadcast")
	}

	select {
	case got := <-c2.send:
		if !bytes.Equal(got, msg) {
			t.Errorf("client2 got %s, want %s", got, msg)
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
