package ssehub

type Client struct {
	Chan chan string

	shutdown chan struct{}
}
