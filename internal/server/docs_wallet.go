package server

import "finance-backend/internal/server/routeinfo"

func walletOpenAPIComponents() map[string]any {
	return map[string]any{
		"schemas": map[string]any{
			"Wallet": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":              map[string]any{"type": "integer"},
					"name":            map[string]any{"type": "string"},
					"opening_balance": map[string]any{"type": "number"},
					"balance":         map[string]any{"type": "number"},
					"created_at":      map[string]any{"type": "string", "format": "date-time"},
					"updated_at":      map[string]any{"type": "string", "format": "date-time"},
				},
			},
			"WalletSummary": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"total_balance": map[string]any{"type": "number"},
					"wallets": map[string]any{
						"type":  "array",
						"items": map[string]any{"$ref": "#/components/schemas/Wallet"},
					},
				},
			},
			"WalletTransfer": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":             map[string]any{"type": "integer"},
					"from_wallet_id": map[string]any{"type": "integer"},
					"to_wallet_id":   map[string]any{"type": "integer"},
					"amount":         map[string]any{"type": "number"},
					"note":           map[string]any{"type": "string"},
					"transfer_date":  map[string]any{"type": "string", "format": "date-time"},
					"created_at":     map[string]any{"type": "string", "format": "date-time"},
					"updated_at":     map[string]any{"type": "string", "format": "date-time"},
				},
			},
			"CreateWalletRequest": map[string]any{
				"type":     "object",
				"required": []string{"name"},
				"properties": map[string]any{
					"name":            map[string]any{"type": "string"},
					"opening_balance": map[string]any{"type": "number"},
				},
			},
			"UpdateWalletRequest": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":            map[string]any{"type": "string"},
					"opening_balance": map[string]any{"type": "number"},
				},
			},
			"CreateWalletTransferRequest": map[string]any{
				"type":     "object",
				"required": []string{"from_wallet_id", "to_wallet_id", "amount"},
				"properties": map[string]any{
					"from_wallet_id": map[string]any{"type": "integer"},
					"to_wallet_id":   map[string]any{"type": "integer"},
					"amount":         map[string]any{"type": "number"},
					"note":           map[string]any{"type": "string"},
					"transfer_date":  map[string]any{"type": "string", "format": "date-time"},
				},
			},
		},
	}
}

func walletRequestBodySchema(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "POST /v1/wallets":
		return jsonRequestBody("#/components/schemas/CreateWalletRequest")
	case "PATCH /v1/wallets/{id}":
		return jsonRequestBody("#/components/schemas/UpdateWalletRequest")
	case "POST /v1/wallet-transfers":
		return jsonRequestBody("#/components/schemas/CreateWalletTransferRequest")
	default:
		return nil
	}
}

func walletResponseSchemas(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "GET /v1/wallets", "GET /v1/wallets/summary", "GET /v1/wallets/{id}", "POST /v1/wallets", "PATCH /v1/wallets/{id}", "DELETE /v1/wallets/{id}", "GET /v1/wallet-transfers", "POST /v1/wallet-transfers":
		return authResponses("#/components/schemas/SuccessResponse")
	default:
		return nil
	}
}
