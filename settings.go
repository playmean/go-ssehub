package ssehub

import "time"

type Settings struct {
	KeepAlive      time.Duration
	DisableLogPage bool
}

type ReceiverSettings struct {
	Method string
}
