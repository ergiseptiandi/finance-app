package httpapi

import (
	"encoding/json"
	"net/http"
	"sort"

	"finance-backend/internal/httpapi/routeinfo"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type routeCatalog map[string]routeinfo.RouteInfo

func newRouteCatalog() routeCatalog {
	return routeCatalog{}
}

func (c routeCatalog) Add(route routeinfo.RouteInfo) {
	c[route.Method+" "+route.Path] = route
}

func (c routeCatalog) List() []routeinfo.RouteInfo {
	routes := make([]routeinfo.RouteInfo, 0, len(c))
	for _, route := range c {
		routes = append(routes, route)
	}

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}

		return routes[i].Path < routes[j].Path
	})

	return routes
}

func registerDocsRoutes(router chi.Router, catalog routeCatalog) {
	router.Get("/routes", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string][]routeinfo.RouteInfo{"routes": catalog.List()})
	})

	router.Get("/openapi.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(buildOpenAPISpec(catalog.List()))
	})

	router.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusTemporaryRedirect)
	})

	router.Handle("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/openapi.json"),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DefaultModelsExpandDepth(-1),
	))
}

func buildOpenAPISpec(routes []routeinfo.RouteInfo) map[string]any {
	paths := map[string]any{}

	for _, route := range routes {
		pathItem, ok := paths[route.Path].(map[string]any)
		if !ok {
			pathItem = map[string]any{}
			paths[route.Path] = pathItem
		}

		operation := map[string]any{
			"summary":     route.Summary,
			"operationId": operationID(route),
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Success",
				},
			},
		}

		if route.Protected {
			operation["security"] = []map[string][]string{
				{"BearerAuth": []string{}},
			}
		}

		if requestBody := requestBodySchema(route); requestBody != nil {
			operation["requestBody"] = requestBody
		}

		if responses := responseSchemas(route); responses != nil {
			operation["responses"] = responses
		}

		pathItem[methodName(route.Method)] = operation
	}

	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "Finance Backend API",
			"version":     "1.0.0",
			"description": "API documentation for the finance backend service.",
		},
		"servers": []map[string]any{
			{"url": "/"},
		},
		"paths": paths,
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"BearerAuth": map[string]any{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
			"schemas": map[string]any{
				"LoginRequest": map[string]any{
					"type": "object",
					"required": []string{
						"email",
						"password",
					},
					"properties": map[string]any{
						"email":       map[string]any{"type": "string", "format": "email"},
						"password":    map[string]any{"type": "string"},
						"device_name": map[string]any{"type": "string"},
					},
				},
				"RefreshRequest": map[string]any{
					"type": "object",
					"required": []string{
						"refresh_token",
					},
					"properties": map[string]any{
						"refresh_token": map[string]any{"type": "string"},
						"device_name":   map[string]any{"type": "string"},
					},
				},
				"LogoutRequest": map[string]any{
					"type": "object",
					"required": []string{
						"refresh_token",
					},
					"properties": map[string]any{
						"refresh_token": map[string]any{"type": "string"},
					},
				},
				"AuthUser": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id":         map[string]any{"type": "integer"},
						"name":       map[string]any{"type": "string"},
						"email":      map[string]any{"type": "string"},
						"created_at": map[string]any{"type": "string", "format": "date-time"},
						"updated_at": map[string]any{"type": "string", "format": "date-time"},
					},
				},
				"TokenPayload": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"access_token":             map[string]any{"type": "string"},
						"access_token_expires_at":  map[string]any{"type": "string", "format": "date-time"},
						"refresh_token":            map[string]any{"type": "string"},
						"refresh_token_expires_at": map[string]any{"type": "string", "format": "date-time"},
						"token_type":               map[string]any{"type": "string"},
					},
				},
				"AuthResult": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"user":  map[string]any{"$ref": "#/components/schemas/AuthUser"},
						"token": map[string]any{"$ref": "#/components/schemas/TokenPayload"},
					},
				},
				"UserProfile": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"user": map[string]any{"$ref": "#/components/schemas/AuthUser"},
					},
				},
				"StatusResponse": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"status": map[string]any{"type": "string"},
					},
				},
				"ErrorResponse": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"error": map[string]any{"type": "string"},
					},
				},
				"RouteInfo": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"method":    map[string]any{"type": "string"},
						"path":      map[string]any{"type": "string"},
						"summary":   map[string]any{"type": "string"},
						"protected": map[string]any{"type": "boolean"},
					},
				},
				"RoutesResponse": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"routes": map[string]any{
							"type":  "array",
							"items": map[string]any{"$ref": "#/components/schemas/RouteInfo"},
						},
					},
				},
			},
		},
	}
}

func operationID(route routeinfo.RouteInfo) string {
	switch route.Method + " " + route.Path {
	case "GET /":
		return "getRootStatus"
	case "GET /health":
		return "getHealthStatus"
	case "GET /routes":
		return "listRoutes"
	case "POST /v1/auth/login":
		return "login"
	case "POST /v1/auth/refresh":
		return "refreshToken"
	case "POST /v1/auth/logout":
		return "logout"
	case "GET /v1/auth/me":
		return "getCurrentUser"
	default:
		return methodName(route.Method) + route.Path
	}
}

func methodName(method string) string {
	switch method {
	case http.MethodGet:
		return "get"
	case http.MethodPost:
		return "post"
	case http.MethodPut:
		return "put"
	case http.MethodPatch:
		return "patch"
	case http.MethodDelete:
		return "delete"
	default:
		return "get"
	}
}

func requestBodySchema(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "POST /v1/auth/login":
		return jsonRequestBody("#/components/schemas/LoginRequest")
	case "POST /v1/auth/refresh":
		return jsonRequestBody("#/components/schemas/RefreshRequest")
	case "POST /v1/auth/logout":
		return jsonRequestBody("#/components/schemas/LogoutRequest")
	default:
		return nil
	}
}

func responseSchemas(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "GET /", "GET /health":
		return successResponse("#/components/schemas/StatusResponse")
	case "GET /routes":
		return successResponse("#/components/schemas/RoutesResponse")
	case "POST /v1/auth/login", "POST /v1/auth/refresh":
		return authResponses("#/components/schemas/AuthResult")
	case "POST /v1/auth/logout":
		return authResponses("#/components/schemas/StatusResponse")
	case "GET /v1/auth/me":
		return authResponses("#/components/schemas/UserProfile")
	default:
		return nil
	}
}

func jsonRequestBody(schemaRef string) map[string]any {
	return map[string]any{
		"required": true,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{
					"$ref": schemaRef,
				},
			},
		},
	}
}

func successResponse(schemaRef string) map[string]any {
	return map[string]any{
		"200": map[string]any{
			"description": "Success",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{"$ref": schemaRef},
				},
			},
		},
	}
}

func authResponses(schemaRef string) map[string]any {
	responses := successResponse(schemaRef)
	responses["400"] = map[string]any{
		"description": "Bad Request",
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"},
			},
		},
	}
	responses["401"] = map[string]any{
		"description": "Unauthorized",
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"},
			},
		},
	}
	return responses
}
