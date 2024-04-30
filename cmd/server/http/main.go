package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ervitis/codenext/internal/input/core"
	http2 "github.com/ervitis/codenext/internal/input/http"
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

func NewHttpServer(port string, endpoints ...core.Endpoint) common.Listener {
	r := http.NewServeMux()
	for _, e := range endpoints {
		var h http.HandlerFunc
		if len(e.Middlewares()) > 0 {
			for i := len(e.Middlewares()) - 1; i >= 0; i-- {
				h = e.Middlewares()[i](e.Handler())
			}
		} else {
			h = e.Handler()
		}
		r.HandleFunc(e.Path(), h)
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

	srv := NewHttpServer(
		"8080",
		[]core.Endpoint{
			http2.NewServerExercisePage(),
			http2.NewServerAnalyzerPage(),
		}...,
	)
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
