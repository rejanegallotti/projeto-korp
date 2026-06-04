package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	serviceUp = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "service_up",
			Help: "Indicates whether the service is up (1) or down (0)",
		},
	)
)

type Response struct {
	Nome    string `json:"nome"`
	Horario string `json:"horario"`
}

func projetoKorpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		requestsTotal.WithLabelValues(r.Method, "/projeto-korp", "405").Inc()
		return
	}

	resp := Response{
		Nome:    "Projeto Korp",
		Horario: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Melhoria 3: tratamento de erro no json.Encode
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("error encoding response: %v", err)
		return
	}

	requestsTotal.WithLabelValues(r.Method, "/projeto-korp", "200").Inc()
	log.Printf("GET /projeto-korp - 200 OK - %s", resp.Horario)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
	requestsTotal.WithLabelValues(r.Method, "/health", "200").Inc()
}

func main() {
	// Melhoria 4: serviceUp vira 0 quando o processo encerrar
	serviceUp.Set(1)
	defer serviceUp.Set(0)

	mux := http.NewServeMux()
	mux.HandleFunc("/projeto-korp", projetoKorpHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// Melhoria 2: timeouts para proteção contra ataques Slowloris
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Melhoria 1: graceful shutdown — inicia o servidor em goroutine separada
	go func() {
		log.Println("http-server-projeto-korp starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Aguarda sinal de encerramento (SIGTERM do Docker, SIGINT do Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Sinal de encerramento recebido. Encerrando graciosamente...")

	// Dá até 30 segundos para requisições em andamento terminarem
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Erro no graceful shutdown: %v", err)
	}

	log.Println("Servidor encerrado com sucesso.")
}
