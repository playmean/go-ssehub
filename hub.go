package ssehub

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Hub struct {
	settings     Settings
	messages     chan Message
	lastMessages []Message
	clients      map[*Client]bool
	mu           sync.Mutex
	shutdown     chan struct{}
	shutdownWg   sync.WaitGroup
	startOnce    sync.Once
	closeOnce    sync.Once
	closed       bool
	closedMutex  sync.RWMutex
}

func NewHub(settings *Settings) *Hub {
	hub := &Hub{
		clients:      make(map[*Client]bool),
		messages:     make(chan Message, 100),
		lastMessages: make([]Message, 0),
		shutdown:     make(chan struct{}),
		settings:     *settings,
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

	hsSent := false

	for _, msg := range hub.lastMessages {
		fmt.Fprintf(w, "data: %s\n\n", msg.Text)

		hsSent = true
	}

	if !hsSent {
		fmt.Fprintf(w, "data: \n\n")
	}

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
	hub.Send(Message{Text: string(p)})

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

			if hub.settings.Retention > 0 && !msg.ping {
				hub.lastMessages = append(hub.lastMessages, msg)

				if len(hub.lastMessages) > hub.settings.Retention {
					hub.lastMessages = hub.lastMessages[1:]
				}
			}

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
			hub.Send(Message{Text: "", ping: true})
		case <-hub.shutdown:
			return
		}
	}
}
