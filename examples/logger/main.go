package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/playmean/go-ssehub"
)

func main() {
	hub := ssehub.NewHub(&ssehub.Settings{
		KeepAlive: 5 * time.Second,
	})

	http.HandleFunc("GET /log", hub.Handler)

	hub.Start()

	logger := hub.NewLogger()

	// Graceful shutdown on signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				logger.Println("Tick")
			case <-stop:
				log.Println("Shutting down...")

				hub.Shutdown()

				os.Exit(0)
			}
		}
	}()

	log.Println("Server started")

	http.ListenAndServe(":8080", nil)
}
