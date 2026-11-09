package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrscorpio/uptimeapp/backend/internal/config"
	"github.com/mrscorpio/uptimeapp/backend/internal/httpapi"
	"github.com/mrscorpio/uptimeapp/backend/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	db, err := store.Open(ctx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()
	server := &http.Server{Addr: cfg.Addr, Handler: httpapi.New(db, cfg).Handler(), ReadHeaderTimeout: 5 * time.Second}
	stop, stopNotify := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopNotify()
	go func() {
		<-stop.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()
	log.Printf("API listening on %s", cfg.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
