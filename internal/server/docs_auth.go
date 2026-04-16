package server

import "finance-backend/internal/server/routeinfo"

func authOpenAPIComponents() map[string]any {
	return map[string]any{
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
		},
	}
}

func authOperationID(route routeinfo.RouteInfo) (string, bool) {
	switch route.Method + " " + route.Path {
	case "POST /v1/auth/login":
		return "login", true
	case "POST /v1/auth/refresh":
		return "refreshToken", true
	case "POST /v1/auth/logout":
		return "logout", true
	case "GET /v1/auth/me":
		return "getCurrentUser", true
	default:
		return "", false
	}
}

func authRequestBodySchema(route routeinfo.RouteInfo) map[string]any {
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

func authResponseSchemas(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
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
