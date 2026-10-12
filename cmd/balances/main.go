package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com.br/devfullcycle/fc-ms-wallet/internal/balances"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(parent context.Context) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	db, err := sql.Open("mysql", env("BALANCES_DB_DSN", "balances:balances@tcp(balances-db:3306)/balances?parseTime=true"))
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	server := &http.Server{
		Addr:              ":3003",
		Handler:           balances.NewHTTPHandler(db),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	failures := make(chan error, 2)
	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		failures <- balances.Consume(ctx, db, env("KAFKA_BROKERS", "kafka:29092"), env("KAFKA_GROUP_ID", "balances-service"))
	}()
	go func() {
		log.Print("Balances API starting on :3003")
		failures <- server.ListenAndServe()
	}()
	select {
	case <-parent.Done():
	case err = <-failures:
	}
	cancel()
	shutdownCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
		log.Printf("HTTP shutdown: %v", shutdownErr)
		server.Close()
	}
	select {
	case <-consumerDone:
	case <-shutdownCtx.Done():
		log.Print("Kafka shutdown timed out")
	}
	if parent.Err() != nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
