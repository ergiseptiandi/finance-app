package server

func openAPIComponents() map[string]any {
	return map[string]any{
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
			"Category": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":         map[string]any{"type": "integer"},
					"name":       map[string]any{"type": "string"},
					"type":       map[string]any{"type": "string", "enum": []string{"income", "expense"}},
					"created_at": map[string]any{"type": "string", "format": "date-time"},
					"updated_at": map[string]any{"type": "string", "format": "date-time"},
				},
			},
			"CategoriesResponse": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"data": map[string]any{
						"type":  "array",
						"items": map[string]any{"$ref": "#/components/schemas/Category"},
					},
				},
			},
			"CreateCategoryRequest": map[string]any{
				"type": "object",
				"required": []string{
					"name",
					"type",
				},
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"type": map[string]any{"type": "string", "enum": []string{"income", "expense"}},
				},
			},
			"UpdateCategoryRequest": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"type": map[string]any{"type": "string", "enum": []string{"income", "expense"}},
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
	}
}
