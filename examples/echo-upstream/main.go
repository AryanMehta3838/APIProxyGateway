package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aryan/apiproxy/internal/testkit"
)

func main() {
	addr := flag.String("addr", ":9091", "listen address (e.g. :9091 or 127.0.0.1:9091)")
	flag.Parse()

	server := &http.Server{
		Addr:    *addr,
		Handler: testkit.EchoHandler(),
	}

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.ListenAndServe()
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("echo upstream listening on %s", *addr)
	select {
	case err := <-serverErrCh:
		if err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "listen on %s: %v\n", *addr, err)
			os.Exit(1)
		}
	case <-stopCtx.Done():
		log.Printf("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "shutdown failed: %v\n", err)
			os.Exit(1)
		}

		if err := <-serverErrCh; err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "shutdown failed: %v\n", err)
			os.Exit(1)
		}

		log.Printf("echo upstream shutdown complete")
	}
}
