package main

import (
	"context"
	"net/http"
	"time"

	"github.com/playmean/go-ssehub"
)

func main() {
	ctx := context.Background()

	hub := ssehub.NewHub(&ssehub.Settings{
		KeepAlive:      5 * time.Second,
		DisableLogPage: true,
	})

	http.HandleFunc("GET /stream", hub.Handler)

	hub.Start()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				hub.Send(ssehub.Message{
					Text: "test",
				})
			case <-ctx.Done():
				return
			}
		}
	}()

	http.ListenAndServe(":8080", nil)
}
