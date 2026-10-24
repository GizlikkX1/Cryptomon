// Package launcher starts every currency microservice and the gateway as
// separate OS processes (re-executing its own binary), restarts them if
// they crash and opens the UI.
package launcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// ExitShutdown is the gateway's exit code when the user chose "Quit" in
// the UI; the launcher then stops everything.
const ExitShutdown = 3

type Child struct {
	Name string
	Args []string
}

type Supervisor struct {
	exe      string
	log      *log.Logger
	children []Child
	quit     chan struct{}
	once     sync.Once
}

func NewSupervisor(exe string, logger *log.Logger, children []Child) *Supervisor {
	return &Supervisor{exe: exe, log: logger, children: children, quit: make(chan struct{})}
}

// Quit asks Run to stop all children.
func (s *Supervisor) Quit() { s.once.Do(func() { close(s.quit) }) }

// Run starts all children and blocks until ctx is cancelled or the UI
// requests a shutdown.
func (s *Supervisor) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup
	for _, c := range s.children {
		wg.Add(1)
		go func(c Child) {
			defer wg.Done()
			s.keepAlive(ctx, c)
		}(c)
	}
	select {
	case <-ctx.Done():
	case <-s.quit:
	}
	cancel()
	wg.Wait()
}

func (s *Supervisor) keepAlive(ctx context.Context, c Child) {
	backoff := time.Second
	for ctx.Err() == nil {
		start := time.Now()
		code, err := s.runOnce(ctx, c)
		if ctx.Err() != nil {
			return
		}
		if code == ExitShutdown {
			s.log.Printf("%s requested shutdown", c.Name)
			s.Quit()
			return
		}
		if time.Since(start) > time.Minute {
			backoff = time.Second
		}
		s.log.Printf("%s exited (%v), restarting in %s", c.Name, err, backoff)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func (s *Supervisor) runOnce(ctx context.Context, c Child) (int, error) {
	cmd := exec.Command(s.exe, c.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	configure(cmd)
	// The child exits when this pipe closes, even if we are killed.
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return -1, err
	}
	if err := cmd.Start(); err != nil {
		return -1, err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		stdin.Close()
		return exitCode(err), err
	case <-ctx.Done():
		stdin.Close()
		interrupt(cmd)
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = cmd.Process.Kill()
			<-done
		}
		return 0, nil
	}
}

func exitCode(err error) int {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	if err == nil {
		return 0
	}
	return -1
}

// PortFree reports whether 127.0.0.1:port can be bound.
func PortFree(port int) bool {
	ln, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// PickPort returns preferred if it is free, otherwise any free port.
func PickPort(preferred int) (int, error) {
	if PortFree(preferred) {
		return preferred, nil
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port, nil
}

// WaitHTTP polls url until it answers 200 or the timeout passes.
func WaitHTTP(ctx context.Context, url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		if resp, err := client.Get(url); err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(300 * time.Millisecond):
		}
	}
	return fmt.Errorf("%s did not respond within %s", url, timeout)
}

// IsCryptomon reports whether a Cryptomon gateway already runs at base.
func IsCryptomon(base string) bool {
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get(base + "/api/meta")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func goos() string { return runtime.GOOS }
