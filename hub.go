package ssehub

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Hub struct {
	settings    Settings
	messages    chan Message
	clients     map[*Client]bool
	mu          sync.Mutex
	shutdown    chan struct{}
	shutdownWg  sync.WaitGroup
	startOnce   sync.Once
	closeOnce   sync.Once
	closed      bool
	closedMutex sync.RWMutex
}

func NewHub(settings *Settings) *Hub {
	hub := &Hub{
		clients:  make(map[*Client]bool),
		messages: make(chan Message, 100),
		shutdown: make(chan struct{}),
		settings: *settings,
	}

	return hub
}

func (hub *Hub) Handler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") != "text/event-stream" {
		if !hub.settings.DisableLogPage {
			hub.pageHandler(w, r)
		} else {
			http.Error(w, "cannot handle plain request", http.StatusBadRequest)
		}

		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)

		return
	}

	client := &Client{
		Chan: make(chan string, 10),

		shutdown: make(chan struct{}),
	}

	hub.mu.Lock()
	hub.clients[client] = true
	hub.mu.Unlock()

	fmt.Fprintf(w, "data: \n\n")
	flusher.Flush()

	defer func() {
		hub.mu.Lock()
		delete(hub.clients, client)
		hub.mu.Unlock()

		close(client.Chan)
	}()

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case <-client.shutdown:
			return
		case msg, ok := <-client.Chan:
			if !ok {
				return
			}

			fmt.Fprintf(w, "data: %s\n\n", msg)

			flusher.Flush()
		}
	}
}

func (hub *Hub) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))

	hub.Send(Message{Text: msg})

	return len(p), nil
}

func (hub *Hub) Send(msg Message) {
	hub.closedMutex.RLock()
	defer hub.closedMutex.RUnlock()

	if hub.closed {
		return
	}

	select {
	case hub.messages <- msg:
	case <-hub.shutdown:
	}
}

func (hub *Hub) Start() {
	hub.startOnce.Do(func() {
		hub.shutdownWg.Add(1)

		go hub.broadcaster()
	})
}

func (hub *Hub) Shutdown() {
	hub.closeOnce.Do(func() {
		hub.closedMutex.Lock()
		hub.closed = true
		hub.closedMutex.Unlock()

		close(hub.shutdown)
		close(hub.messages)
		hub.shutdownWg.Wait()

		hub.mu.Lock()
		for client := range hub.clients {
			close(client.shutdown)
		}
		hub.clients = nil
		hub.mu.Unlock()
	})
}

func (hub *Hub) broadcaster() {
	defer hub.shutdownWg.Done()

	ticker := time.NewTicker(hub.settings.KeepAlive)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-hub.messages:
			if !ok {
				return
			}

			hub.mu.Lock()
			for client := range hub.clients {
				select {
				case client.Chan <- msg.Text:
				default:
					close(client.Chan)
					delete(hub.clients, client)
				}
			}
			hub.mu.Unlock()
		case <-ticker.C:
			hub.Send(Message{Text: ""})
		case <-hub.shutdown:
			return
		}
	}
}
