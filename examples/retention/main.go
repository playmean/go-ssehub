package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/playmean/go-ssehub"
)

func main() {
	ctx := context.Background()

	hub := ssehub.NewHub(&ssehub.Settings{
		KeepAlive: 5 * time.Second,
		Retention: 5,
	})

	client := ssehub.NewReceiver(ctx, "http://localhost:8080/log", &ssehub.ReceiverSettings{
		LinesBufferSize: 10,
	})

	http.HandleFunc("GET /log", hub.Handler)

	hub.Start()

	go http.ListenAndServe(":8080", nil)

	go func() {
		time.Sleep(10 * time.Second)

		err := client.Connect()
		if err != nil {
			panic(err)
		}

		for {
			line, err := client.Next()
			if err != nil {
				panic(err)
			}

			fmt.Println(line)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hub.Send(ssehub.Message{
				Text: time.Now().String(),
			})
		case <-ctx.Done():
			return
		}
	}
}
