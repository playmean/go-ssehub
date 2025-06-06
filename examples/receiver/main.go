package main

import (
	"context"
	"fmt"
	"time"

	"github.com/playmean/go-ssehub"
)

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	client := ssehub.NewReceiver(ctx, "http://localhost:8080/stream", &ssehub.ReceiverSettings{
		LinesBufferSize: 3,
	})

	err := client.Connect()
	if err != nil {
		panic(err)
	}

	time.Sleep(3 * time.Second)

	for {
		line, err := client.Next()
		if err != nil {
			fmt.Println(err)

			break
		}

		fmt.Println(line)
	}

	client.Shutdown()
}
