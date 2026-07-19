// Command hazinadb is the composition root of the application.
//
// This is the ONLY place that knows about concrete types. It picks the
// adapters, wires them to the core, and starts the server. Swapping the
// storage engine or the transport is a change here and nowhere else — that is
// the payoff of the ports-and-adapters layout.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/luckybaraka/hazinadb/internal/adapter/inbound/rest"
	"github.com/luckybaraka/hazinadb/internal/adapter/outbound/memory"
	"github.com/luckybaraka/hazinadb/internal/core/service"
)

func main() {
	addr := flag.String("addr", ":8000", "address the HTTP server listens on")
	flag.Parse()

	// Wire the hexagon: engine (driven adapter) -> service (core) -> server
	// (driving adapter). Dependencies point inward at the core.
	engine := memory.NewEngine()
	defer engine.Close()

	store := service.NewKeyValue(engine)
	server := rest.NewServer(*addr, store)

	// Run the server in the background so main can wait for a shutdown signal.
	go func() {
		log.Printf("HazinaDB key-value store listening on %s", *addr)
		if err := server.ListenAndServe(); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	// Block until the process is asked to stop (Ctrl+C / SIGTERM), then drain
	// in-flight requests within a bounded window before exiting.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	<-ctx.Done()

	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
}
