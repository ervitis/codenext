package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ervitis/codenext/cmd/server/common"
)

type httpServer struct {
	srv *http.Server
}

func (h httpServer) Shutdown(ctx context.Context) error {
	return h.srv.Shutdown(ctx)
}

func (h httpServer) ListenAndServe() error {
	if err := h.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}

type endpoint struct {
	method string
	fn     http.HandlerFunc
	mids   []func(http.HandlerFunc) http.HandlerFunc
	path   string
}

func NewHttpServer(port string, endpoints ...endpoint) common.Listener {
	r := http.NewServeMux()
	for _, e := range endpoints {
		var h http.HandlerFunc
		for i := len(e.mids) - 1; i >= 0; i-- {
			h = e.mids[i](e.fn)
		}
		r.HandleFunc(e.path, h)
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       time.Minute,
	}

	return &httpServer{
		srv: srv,
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer stop()

	end := make(chan struct{}, 1)

	srv := NewHttpServer("8080", []endpoint{}...)
	go func() {
		log.Printf("serving http on port: %v", "8080")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("http server: %v", err)
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				withCancel, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				cancel()
				_ = srv.Shutdown(withCancel)
				end <- struct{}{}
			}
		}
	}()
	<-end
}
