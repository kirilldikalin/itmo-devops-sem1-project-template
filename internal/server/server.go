package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"project_sem/internal/api"
	"project_sem/platform/config"
	"project_sem/platform/storage"
)

type Application interface {
	Run()
}

type application struct {
	server   *http.Server
	database storage.Repository
	quit     chan os.Signal
	wg       *sync.WaitGroup
}

func New(cfg config.Settings) (Application, error) {
	quit := make(chan os.Signal, 1)
	wg := &sync.WaitGroup{}

	repo, err := storage.NewRepository(cfg.Database)
	if err != nil {
		log.Printf("failed to create repository: %v", err)
		close(quit)
		return nil, fmt.Errorf("repository initialization failed: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v0/prices", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			api.PostPrices(repo, cfg.Server.MaxFileSize)(w, r)
		case http.MethodGet:
			api.GetPrices(repo)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	serverInstance := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return &application{
		server:   serverInstance,
		database: repo,
		quit:     quit,
		wg:       wg,
	}, nil
}

func (a *application) Run() {
	signal.Notify(a.quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Printf("Starting server on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server failed: %v", err)
			a.quit <- syscall.SIGTERM
		}
	}()

	<-a.quit
	log.Println("Shutting down gracefully...")

	const shutdownTimeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.server.Shutdown(ctx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		} else {
			log.Println("Server shut down gracefully.")
		}
	}()

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.database.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database connection closed successfully.")
		}
	}()

	a.wg.Wait()

	close(a.quit)
	log.Println("All resources released. Server shutdown complete.")
}