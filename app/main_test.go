package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProjetoKorpHandler(t *testing.T) {
	// Cria uma requisição GET para /projeto-korp
	req := httptest.NewRequest(http.MethodGet, "/projeto-korp", nil)
	w := httptest.NewRecorder()

	// Chama o handler diretamente
	projetoKorpHandler(w, req)

	// Verifica status HTTP 200
	if w.Code != http.StatusOK {
		t.Errorf("esperado status 200, recebido %d", w.Code)
	}

	// Verifica Content-Type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("esperado Content-Type application/json, recebido %s", contentType)
	}

	// Decodifica e verifica o JSON retornado
	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("erro ao decodificar resposta JSON: %v", err)
	}

	// Verifica campo nome
	if resp.Nome != "Projeto Korp" {
		t.Errorf("esperado nome 'Projeto Korp', recebido '%s'", resp.Nome)
	}

	// Verifica que horario está preenchido e é UTC válido
	if resp.Horario == "" {
		t.Error("campo horario está vazio")
	}

	_, err := time.Parse(time.RFC3339, resp.Horario)
	if err != nil {
		t.Errorf("horario não está no formato RFC3339: %s", resp.Horario)
	}
}

func TestProjetoKorpHandler_MethodNotAllowed(t *testing.T) {
	// Testa que POST retorna 405
	req := httptest.NewRequest(http.MethodPost, "/projeto-korp", nil)
	w := httptest.NewRecorder()

	projetoKorpHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("esperado status 405, recebido %d", w.Code)
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("esperado status 200, recebido %d", w.Code)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("erro ao decodificar resposta JSON: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("esperado status 'ok', recebido '%s'", result["status"])
	}
}
