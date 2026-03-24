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

	"github.com/aryan/apiproxy/internal/admin"
	"github.com/aryan/apiproxy/internal/config"
	"github.com/aryan/apiproxy/internal/router"
)

func main() {
	configPath := flag.String("config", "configs/gateway.dev.yaml", "path to gateway config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "startup failed: %v\n", err)
		os.Exit(1)
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	handler, err := router.New(cfg, admin.NewRouter())
	if err != nil {
		fmt.Fprintf(os.Stderr, "startup failed: %v\n", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.ListenAndServe()
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("gateway listening on %s", addr)
	select {
	case err := <-serverErrCh:
		if err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "startup failed: listen on %s: %v\n", addr, err)
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

		log.Printf("gateway shutdown complete")
	}
}
