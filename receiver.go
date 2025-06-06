package ssehub

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Receiver struct {
	ctx      context.Context
	url      string
	resp     *http.Response
	out      io.Writer
	shutdown bool
	ch       chan string
	chTick   chan string
	settings *ReceiverSettings

	C chan string
}

func NewReceiver(ctx context.Context, url string, settings *ReceiverSettings) *Receiver {
	if settings == nil {
		settings = &ReceiverSettings{}
	}

	if settings.Method == "" {
		settings.Method = "GET"
	}

	r := Receiver{
		ctx:      ctx,
		url:      url,
		resp:     nil,
		out:      nil,
		shutdown: true,
		ch:       make(chan string),
		chTick:   make(chan string, settings.LinesBufferSize),
		settings: settings,

		C: make(chan string, settings.LinesBufferSize),
	}

	return &r
}

func (r *Receiver) SetOutput(out io.Writer) {
	r.out = out
}

func (r *Receiver) Connect() error {
	err := r.makeRequest()
	if err != nil {
		return err
	}

	go r.receiveLoop()
	go r.scanLoop()

	return nil
}

func (r *Receiver) ConnectSync() error {
	err := r.makeRequest()
	if err != nil {
		return err
	}

	go r.receiveLoop()

	r.scanLoop()

	return nil
}

func (r *Receiver) Shutdown() {
	r.shutdown = true
}

func (r *Receiver) Next() (string, error) {
	select {
	case line := <-r.chTick:
		return line, nil
	case <-r.ctx.Done():
		return "", fmt.Errorf("context done")
	}
}

func (r *Receiver) makeRequest() error {
	var err error

	req, err := http.NewRequest(r.settings.Method, r.url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "text/event-stream")

	r.resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if r.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response code: %s", r.resp.Status)
	}

	r.shutdown = false

	return nil
}

func (r *Receiver) scanLoop() {
	scanner := bufio.NewScanner(r.resp.Body)

	for !r.shutdown && scanner.Scan() {
		message := scanner.Text()

		if !strings.HasPrefix(message, "data: ") {
			continue
		}

		line := strings.TrimPrefix(message, "data: ")

		if line == "" {
			continue
		}

		r.ch <- line
	}

	r.resp.Body.Close()
}

func (r *Receiver) receiveLoop() {
	for {
		select {
		case line := <-r.ch:
			if r.out != nil {
				r.out.Write([]byte(line))
			}

			select {
			case r.chTick <- line:
			default:
			}

			select {
			case r.C <- line:
			default:
			}

		case <-r.ctx.Done():
			r.shutdown = true

			return
		}
	}
}
