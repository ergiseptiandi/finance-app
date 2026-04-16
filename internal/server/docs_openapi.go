package server

import (
	"net/http"

	"finance-backend/internal/server/routeinfo"
)

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
		"paths":      paths,
		"components": openAPIComponents(),
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
	case "POST /v1/categories":
		return jsonRequestBody("#/components/schemas/CreateCategoryRequest")
	case "PATCH /v1/categories/{id}":
		return jsonRequestBody("#/components/schemas/UpdateCategoryRequest")
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
	case "GET /v1/categories":
		return authResponses("#/components/schemas/CategoriesResponse")
	case "POST /v1/categories", "PATCH /v1/categories/{id}":
		return authResponses("#/components/schemas/Category")
	case "DELETE /v1/categories/{id}":
		return authResponses("#/components/schemas/StatusResponse")
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
