package server

import "finance-backend/internal/server/routeinfo"

func transactionOpenAPIComponents() map[string]any {
	return map[string]any{
		"schemas": map[string]any{
			"TransactionType": map[string]any{
				"type": "string",
				"enum": []string{"income", "expense"},
			},
			"Transaction": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":          map[string]any{"type": "integer"},
					"user_id":     map[string]any{"type": "integer"},
					"type":        map[string]any{"$ref": "#/components/schemas/TransactionType"},
					"category":    map[string]any{"type": "string"},
					"amount":      map[string]any{"type": "number"},
					"date":        map[string]any{"type": "string", "format": "date"},
					"description": map[string]any{"type": "string"},
					"created_at":  map[string]any{"type": "string", "format": "date-time"},
					"updated_at":  map[string]any{"type": "string", "format": "date-time"},
				},
			},
			"CreateTransactionRequest": map[string]any{
				"type": "object",
				"required": []string{
					"type",
					"category",
					"amount",
					"date",
				},
				"properties": map[string]any{
					"type":        map[string]any{"$ref": "#/components/schemas/TransactionType"},
					"category":    map[string]any{"type": "string"},
					"amount":      map[string]any{"type": "number"},
					"date":        map[string]any{"type": "string", "format": "date"},
					"description": map[string]any{"type": "string"},
				},
			},
			"UpdateTransactionRequest": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"type":        map[string]any{"$ref": "#/components/schemas/TransactionType"},
					"category":    map[string]any{"type": "string"},
					"amount":      map[string]any{"type": "number"},
					"date":        map[string]any{"type": "string", "format": "date"},
					"description": map[string]any{"type": "string"},
				},
			},
			"TransactionSummary": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"total_income":  map[string]any{"type": "number"},
					"total_expense": map[string]any{"type": "number"},
					"balance":       map[string]any{"type": "number"},
				},
			},
			"TransactionListResponse": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"data": map[string]any{
						"type":  "array",
						"items": map[string]any{"$ref": "#/components/schemas/Transaction"},
					},
					"total":       map[string]any{"type": "integer"},
					"page":        map[string]any{"type": "integer"},
					"per_page":    map[string]any{"type": "integer"},
					"total_pages": map[string]any{"type": "integer"},
				},
			},
		},
	}
}

func transactionRequestBodySchema(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "POST /v1/transactions":
		return jsonRequestBody("#/components/schemas/CreateTransactionRequest")
	case "PATCH /v1/transactions/{id}":
		return jsonRequestBody("#/components/schemas/UpdateTransactionRequest")
	default:
		return nil
	}
}

func transactionResponseSchemas(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "GET /v1/transactions":
		return authResponses("#/components/schemas/SuccessResponse")
	case "GET /v1/transactions/summary":
		return authResponses("#/components/schemas/SuccessResponse")
	case "POST /v1/transactions", "GET /v1/transactions/{id}", "PATCH /v1/transactions/{id}":
		return authResponses("#/components/schemas/SuccessResponse")
	case "DELETE /v1/transactions/{id}":
		return authResponses("#/components/schemas/SuccessResponse")
	default:
		return nil
	}
}
