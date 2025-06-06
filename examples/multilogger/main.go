package main

import (
	"context"
	"io"
	"log"
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
		KeepAlive: 5 * time.Second,
	})

	http.HandleFunc("GET /log", hub.Handler)

	hub.Start()

	log.SetOutput(io.MultiWriter(hub, os.Stdout))

	go http.ListenAndServe(":8080", nil)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Tick")
		case <-ctx.Done():
			log.Println("Shutting down...")

			hub.Shutdown()

			return
		}
	}
}
