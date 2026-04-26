package server

import "finance-backend/internal/server/routeinfo"

func debtOpenAPIComponents() map[string]any {
	return map[string]any{
		"schemas": map[string]any{
			"DebtStatus": map[string]any{
				"type": "string",
				"enum": []string{"pending", "paid", "overdue"},
			},
			"Debt": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":                  map[string]any{"type": "integer"},
					"user_id":             map[string]any{"type": "integer"},
					"name":                map[string]any{"type": "string"},
					"total_amount":        map[string]any{"type": "number"},
					"monthly_installment": map[string]any{"type": "number"},
					"due_date":            map[string]any{"type": "string", "format": "date-time"},
					"paid_amount":         map[string]any{"type": "number"},
					"remaining_amount":    map[string]any{"type": "number"},
					"status":              map[string]any{"$ref": "#/components/schemas/DebtStatus"},
					"created_at":          map[string]any{"type": "string", "format": "date-time"},
					"updated_at":          map[string]any{"type": "string", "format": "date-time"},
				},
			},
			"Installment": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":             map[string]any{"type": "integer"},
					"debt_id":        map[string]any{"type": "integer"},
					"installment_no": map[string]any{"type": "integer"},
					"due_date":       map[string]any{"type": "string", "format": "date-time"},
					"amount":         map[string]any{"type": "number"},
					"status":         map[string]any{"$ref": "#/components/schemas/DebtStatus"},
					"paid_at":        map[string]any{"type": "string", "format": "date-time", "nullable": true},
					"created_at":     map[string]any{"type": "string", "format": "date-time"},
					"updated_at":     map[string]any{"type": "string", "format": "date-time"},
				},
			},
			"Payment": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":             map[string]any{"type": "integer"},
					"debt_id":        map[string]any{"type": "integer"},
					"wallet_id":      map[string]any{"type": "integer"},
					"installment_id": map[string]any{"type": "integer", "nullable": true},
					"amount":         map[string]any{"type": "number"},
					"payment_date":   map[string]any{"type": "string", "format": "date-time"},
					"proof_image":    map[string]any{"type": "string", "format": "uri"},
					"created_at":     map[string]any{"type": "string", "format": "date-time"},
					"updated_at":     map[string]any{"type": "string", "format": "date-time"},
				},
			},
			"DebtDetail": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":                   map[string]any{"type": "integer"},
					"user_id":              map[string]any{"type": "integer"},
					"name":                 map[string]any{"type": "string"},
					"total_amount":         map[string]any{"type": "number"},
					"monthly_installment":  map[string]any{"type": "number"},
					"due_date":             map[string]any{"type": "string", "format": "date-time"},
					"paid_amount":          map[string]any{"type": "number"},
					"remaining_amount":     map[string]any{"type": "number"},
					"status":               map[string]any{"$ref": "#/components/schemas/DebtStatus"},
					"paid_installments":    map[string]any{"type": "integer"},
					"unpaid_installments":  map[string]any{"type": "integer"},
					"overdue_installments": map[string]any{"type": "integer"},
					"installments":         map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/Installment"}},
					"payments":             map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/Payment"}},
				},
			},
			"CreateDebtRequest": map[string]any{
				"type":     "object",
				"required": []string{"name", "total_amount", "monthly_installment", "due_date"},
				"properties": map[string]any{
					"name":                map[string]any{"type": "string"},
					"total_amount":        map[string]any{"type": "number"},
					"monthly_installment": map[string]any{"type": "number"},
					"due_date":            map[string]any{"type": "string", "format": "date-time"},
				},
			},
			"UpdateDebtRequest": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":                map[string]any{"type": "string"},
					"total_amount":        map[string]any{"type": "number"},
					"monthly_installment": map[string]any{"type": "number"},
					"due_date":            map[string]any{"type": "string", "format": "date-time"},
				},
			},
			"CreatePaymentRequest": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"wallet_id":    map[string]any{"type": "integer"},
					"amount":       map[string]any{"type": "number"},
					"payment_date": map[string]any{"type": "string", "format": "date-time"},
					"proof_image":  map[string]any{"type": "string", "format": "binary"},
				},
			},
			"UpdatePaymentRequest": map[string]any{
				"type": "object",
				"description": "All fields are optional. Provide proof_image to replace the stored payment proof.",
				"properties": map[string]any{
					"wallet_id":    map[string]any{"type": "integer"},
					"amount":       map[string]any{"type": "number"},
					"payment_date": map[string]any{"type": "string", "format": "date-time"},
					"proof_image":  map[string]any{"type": "string", "format": "binary"},
				},
			},
			"MarkInstallmentPaidRequest": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"paid_at": map[string]any{"type": "string", "format": "date-time"},
				},
			},
		},
	}
}

func debtRequestBodySchema(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "POST /v1/debts":
		return jsonRequestBody("#/components/schemas/CreateDebtRequest")
	case "PATCH /v1/debts/{id}":
		return jsonRequestBody("#/components/schemas/UpdateDebtRequest")
	case "POST /v1/debts/{id}/payments":
		return map[string]any{
			"required": true,
			"content": map[string]any{
				"multipart/form-data": map[string]any{
					"schema": map[string]any{"$ref": "#/components/schemas/CreatePaymentRequest"},
				},
			},
		}
	case "PATCH /v1/debts/{id}/payments/{paymentId}":
		return map[string]any{
			"required": true,
			"content": map[string]any{
				"multipart/form-data": map[string]any{
					"schema": map[string]any{"$ref": "#/components/schemas/UpdatePaymentRequest"},
				},
			},
		}
	case "PATCH /v1/debts/{id}/installments/{installmentId}/paid":
		return jsonRequestBody("#/components/schemas/MarkInstallmentPaidRequest")
	default:
		return nil
	}
}

func debtResponseSchemas(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "POST /v1/debts", "GET /v1/debts/{id}":
		return authResponses("#/components/schemas/SuccessResponse")
	case "GET /v1/debts":
		return debtAuthResponsesWithSuccessDescription(
			"#/components/schemas/SuccessResponse",
			"Success (returns empty array in Data when no debts are available)",
		)
	case "PATCH /v1/debts/{id}", "DELETE /v1/debts/{id}":
		return authResponses("#/components/schemas/SuccessResponse")
	case "GET /v1/debts/{id}/payments", "GET /v1/debts/{id}/installments":
		return debtAuthResponsesWithSuccessDescription(
			"#/components/schemas/SuccessResponse",
			"Success (returns empty array in Data when no records are available)",
		)
	case "POST /v1/debts/{id}/payments", "PATCH /v1/debts/{id}/payments/{paymentId}", "PATCH /v1/debts/{id}/installments/{installmentId}/paid":
		return authResponses("#/components/schemas/SuccessResponse")
	default:
		return nil
	}
}

func debtAuthResponsesWithSuccessDescription(schemaRef, description string) map[string]any {
	responses := authResponses(schemaRef)
	if success, ok := responses["200"].(map[string]any); ok {
		success["description"] = description
	}
	return responses
}
