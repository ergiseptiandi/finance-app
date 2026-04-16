package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-backend/internal/httpapi/routeinfo"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	NewRouter(nil, nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	expected := "{\"status\":\"ok\"}\n"
	if rec.Body.String() != expected {
		t.Fatalf("expected body %q, got %q", expected, rec.Body.String())
	}
}

func TestRoutesEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/routes", nil)
	rec := httptest.NewRecorder()

	NewRouter(nil, nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload struct {
		Routes []routeinfo.RouteInfo `json:"routes"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode routes response: %v", err)
	}

	if len(payload.Routes) < 3 {
		t.Fatalf("expected at least 3 routes, got %d", len(payload.Routes))
	}
}

func TestOpenAPIEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()

	NewRouter(nil, nil).ServeHTTP(rec, req)

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

	NewRouter(nil, nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}

	if location := rec.Header().Get("Location"); location != "/docs/index.html" {
		t.Fatalf("expected redirect to /docs/index.html, got %q", location)
	}
}
