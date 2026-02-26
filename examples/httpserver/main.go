package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := run()

	go func() {
		log.Printf("listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down test server")
	if err := server.Close(); err != nil {
		log.Printf("server close failed: %v", err)
	}
}

func run() *http.Server {
	var counter atomic.Int64

	mux := http.NewServeMux()
	mux.HandleFunc("POST /increment", func(w http.ResponseWriter, _ *http.Request) {
		counter.Add(1)
		writeCounter(w, counter.Load())
	})
	mux.HandleFunc("POST /decrement", func(w http.ResponseWriter, _ *http.Request) {
		counter.Add(-1)
		writeCounter(w, counter.Load())
	})
	mux.HandleFunc("POST /zero", func(w http.ResponseWriter, _ *http.Request) {
		counter.Store(0)
		writeCounter(w, counter.Load())
	})
	mux.HandleFunc("POST /add/{n}", func(w http.ResponseWriter, r *http.Request) {
		n, err := strconv.ParseInt(r.PathValue("n"), 10, 64)
		if err != nil {
			http.Error(w, "invalid n", http.StatusBadRequest)
			return
		}
		counter.Add(n)
		writeCounter(w, counter.Load())
	})
	mux.HandleFunc("POST /sub/{n}", func(w http.ResponseWriter, r *http.Request) {
		n, err := strconv.ParseInt(r.PathValue("n"), 10, 64)
		if err != nil {
			http.Error(w, "invalid n", http.StatusBadRequest)
			return
		}
		counter.Add(-n)
		writeCounter(w, counter.Load())
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
}

func writeCounter(w http.ResponseWriter, count int64) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Count int64 `json:"count"`
	}{Count: count})
}
