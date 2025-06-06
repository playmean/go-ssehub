package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/playmean/go-ssehub"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	hub := ssehub.NewHub(&ssehub.Settings{
		KeepAlive:      5 * time.Second,
		DisableLogPage: true,
	})

	http.HandleFunc("GET /stream", hub.Handler)

	hub.Start()

	go http.ListenAndServe(":8080", nil)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hub.Send(ssehub.Message{
				Text: time.Now().String(),
			})
		case <-ctx.Done():
			hub.Shutdown()

			return
		}
	}
}
