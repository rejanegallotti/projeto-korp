package main

import (
	"encoding/json"
	"log"
	"net/http"
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
	json.NewEncoder(w).Encode(resp)

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
	// Mark service as up
	serviceUp.Set(1)

	mux := http.NewServeMux()
	mux.HandleFunc("/projeto-korp", projetoKorpHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/metrics", promhttp.Handler())

	log.Println("http-server-projeto-korp starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		serviceUp.Set(0)
		log.Fatalf("Server failed: %v", err)
	}
}
