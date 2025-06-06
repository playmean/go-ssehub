package ssehub

import "time"

type Settings struct {
	KeepAlive      time.Duration
	DisableLogPage bool
	Retention      int
}

type ReceiverSettings struct {
	Method          string
	LinesBufferSize int
}
