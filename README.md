# SSEHub

📡 **SSEHub** is a lightweight Go library for streaming logs and events to browser clients using Server-Sent Events (SSE). It supports safe concurrency, integration with the standard logger and graceful shutdown.

## ✨ Features

- 📤 Real-time streaming via Server-Sent Events (SSE)
- 🧩 Integration with Go’s standard `log.Logger` and `http.Handler`
- 🔁 Built-in keep-alive pings to prevent client timeouts
- 💥 Safe message sending after shutdown
- 🧼 Graceful shutdown support

## 🚀 Installation

```bash
go get github.com/playmean/go-ssehub
```

## 🧪 Example Usage

```go
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
```

Then open `http://localhost:8080/log` to view realtime logs.

## 🌐 Custom Client (Browser Example)

```js
const source = new EventSource("/stream");

source.onmessage = (event) => {
    console.log("SSE:", event.data);
};

source.onerror = (err) => {
    console.error("SSE connection error", err);
};
```

## ⚙️ Why SSEHub?

- 🖥️ Watch live logs from your app directly in the browser
- 🔧 No third-party brokers or dependencies required
- 🧘 Safe, simple and production-ready

## 📄 License

MIT
