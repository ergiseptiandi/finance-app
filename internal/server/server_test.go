package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-backend/internal/server/routeinfo"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, "").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload struct {
		Status  string `json:"Status"`
		Message string `json:"Message"`
		Data    struct {
			Status string `json:"status"`
		} `json:"Data"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}

	if payload.Status != "200" {
		t.Fatalf("expected status field 200, got %q", payload.Status)
	}
	if payload.Message != "Success Get" {
		t.Fatalf("expected message Success Get, got %q", payload.Message)
	}
	if payload.Data.Status != "ok" {
		t.Fatalf("expected data status ok, got %q", payload.Data.Status)
	}
}

func TestRoutesEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/routes", nil)
	rec := httptest.NewRecorder()

	NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, "").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload struct {
		Status  string `json:"Status"`
		Message string `json:"Message"`
		Data    struct {
			Routes []routeinfo.RouteInfo `json:"routes"`
		} `json:"Data"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode routes response: %v", err)
	}

	if payload.Status != "200" {
		t.Fatalf("expected status field 200, got %q", payload.Status)
	}
	if len(payload.Data.Routes) < 3 {
		t.Fatalf("expected at least 3 routes, got %d", len(payload.Data.Routes))
	}
}

func TestOpenAPIEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()

	NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, "").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode openapi response: %v", err)
	}

	if payload["openapi"] != "3.0.3" {
		t.Fatalf("expected openapi version 3.0.3, got %v", payload["openapi"])
	}
}

func TestDocsRedirect(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()

	NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, "").ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}

	if location := rec.Header().Get("Location"); location != "/docs/index.html" {
		t.Fatalf("expected redirect to /docs/index.html, got %q", location)
	}
}
