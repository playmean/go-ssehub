package main

import (
	"context"
	"fmt"
	"time"

	"github.com/playmean/go-ssehub"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := ssehub.NewReceiver(ctx, "http://localhost:8080/stream", &ssehub.ReceiverSettings{
		LinesBufferSize: 3,
	})

	err := client.Connect()
	if err != nil {
		panic(err)
	}

	time.Sleep(3 * time.Second)

	for {
		line, err, done := client.Next()
		if done {
			break
		}
		if err != nil {
			panic(err)
		}

		fmt.Println(line)
	}

	client.Shutdown()
}
