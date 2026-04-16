package server

import "finance-backend/internal/server/routeinfo"

func categoryOpenAPIComponents() map[string]any {
	return map[string]any{
		"schemas": map[string]any{
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
		},
	}
}

func categoryRequestBodySchema(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "POST /v1/categories":
		return jsonRequestBody("#/components/schemas/CreateCategoryRequest")
	case "PATCH /v1/categories/{id}":
		return jsonRequestBody("#/components/schemas/UpdateCategoryRequest")
	default:
		return nil
	}
}

func categoryResponseSchemas(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "GET /v1/categories":
		return authResponses("#/components/schemas/SuccessResponse")
	case "POST /v1/categories", "PATCH /v1/categories/{id}":
		return authResponses("#/components/schemas/SuccessResponse")
	case "DELETE /v1/categories/{id}":
		return authResponses("#/components/schemas/SuccessResponse")
	default:
		return nil
	}
}
