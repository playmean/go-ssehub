package main

import (
	"context"
	"os"

	"github.com/playmean/go-ssehub"
)

func main() {
	ctx := context.Background()

	client := ssehub.NewReceiver(ctx, "http://localhost:8080/stream", &ssehub.ReceiverSettings{})
	client.SetOutput(os.Stdout)

	err := client.ConnectSync()
	if err != nil {
		panic(err)
	}
}
