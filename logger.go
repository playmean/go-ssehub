package ssehub

import "log"

type Logger struct {
	Hub *Hub
}

func (hub *Hub) NewLogger() *log.Logger {
	return log.New(&Logger{Hub: hub}, "", log.LstdFlags)
}

func (tl *Logger) Write(p []byte) (n int, err error) {
	return tl.Hub.Write(p)
}
