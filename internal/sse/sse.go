// Package sse implements Server-Sent Events: a broadcast broker for
// servers and a reconnecting client used by the gateway to follow the
// currency services.
package sse

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Broker fans published events out to every connected HTTP client.
type Broker struct {
	mu      sync.Mutex
	clients map[chan []byte]struct{}
	last    []byte // replayed to new clients
}

func NewBroker() *Broker {
	return &Broker{clients: map[chan []byte]struct{}{}}
}

// Format renders one SSE frame.
func Format(event string, data []byte) []byte {
	var b strings.Builder
	if event != "" {
		b.WriteString("event: ")
		b.WriteString(event)
		b.WriteByte('\n')
	}
	for _, line := range strings.Split(string(data), "\n") {
		b.WriteString("data: ")
		b.WriteString(line)
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
	return []byte(b.String())
}

// Publish sends an event to every client; slow clients drop frames rather
// than block the publisher.
func (b *Broker) Publish(event string, data []byte) {
	frame := Format(event, data)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.last = frame
	for ch := range b.clients {
		select {
		case ch <- frame:
		default:
		}
	}
}

func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	PrepareHeaders(w)
	ch := make(chan []byte, 32)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	last := b.last
	b.mu.Unlock()
	defer func() {
		b.mu.Lock()
		delete(b.clients, ch)
		b.mu.Unlock()
	}()
	if last != nil {
		_, _ = w.Write(last)
	}
	flusher.Flush()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case frame := <-ch:
			if _, err := w.Write(frame); err != nil {
				return
			}
			flusher.Flush()
		case <-heartbeat.C:
			if _, err := w.Write([]byte(": ping\n\n")); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func PrepareHeaders(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
}

// streamClient has no overall timeout: streams are long-lived.
var streamClient = &http.Client{}

// Subscribe follows an SSE endpoint until ctx is cancelled, reconnecting
// with backoff. onConnect is called on every (re)connection with the error
// state; handler gets every event.
func Subscribe(ctx context.Context, url string, onState func(connected bool, err error), handler func(event string, data []byte)) {
	backoff := time.Second
	for ctx.Err() == nil {
		err := stream(ctx, url, func() {
			backoff = time.Second
			onState(true, nil)
		}, handler)
		if ctx.Err() != nil {
			return
		}
		onState(false, err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < 10*time.Second {
			backoff *= 2
		}
	}
}

func stream(ctx context.Context, url string, connected func(), handler func(string, []byte)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")
	resp, err := streamClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	connected()

	// A watchdog closes the stream if nothing (not even a heartbeat)
	// arrives for 45s, so dead peers are detected.
	alive := make(chan struct{}, 1)
	done := make(chan struct{})
	defer close(done)
	go func() {
		t := time.NewTimer(45 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-alive:
				if !t.Stop() {
					<-t.C
				}
				t.Reset(45 * time.Second)
			case <-t.C:
				resp.Body.Close()
				return
			}
		}
	}()

	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64*1024), 8<<20)
	var event string
	var data strings.Builder
	for sc.Scan() {
		select {
		case alive <- struct{}{}:
		default:
		}
		line := sc.Text()
		switch {
		case line == "":
			if data.Len() > 0 {
				handler(event, []byte(data.String()))
			}
			event = ""
			data.Reset()
		case strings.HasPrefix(line, ":"):
		case strings.HasPrefix(line, "event:"):
			event = strings.TrimSpace(line[6:])
		case strings.HasPrefix(line, "data:"):
			if data.Len() > 0 {
				data.WriteByte('\n')
			}
			data.WriteString(strings.TrimPrefix(line[5:], " "))
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return fmt.Errorf("stream closed")
}
