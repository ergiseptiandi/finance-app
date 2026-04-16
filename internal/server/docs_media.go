package server

import "finance-backend/internal/server/routeinfo"

func mediaOpenAPIComponents() map[string]any {
	return map[string]any{
		"schemas": map[string]any{
			"MediaUploadResult": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":      map[string]any{"type": "string"},
					"url":       map[string]any{"type": "string"},
					"dir":       map[string]any{"type": "string"},
					"file_name": map[string]any{"type": "string"},
				},
			},
			"MediaUrlResponse": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{"type": "string"},
					"url":  map[string]any{"type": "string"},
				},
			},
		},
	}
}

func mediaRequestBodySchema(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "POST /v1/media/upload":
		return map[string]any{
			"required": true,
			"content": map[string]any{
				"multipart/form-data": map[string]any{
					"schema": map[string]any{
						"type":     "object",
						"required": []string{"file"},
						"properties": map[string]any{
							"file": map[string]any{"type": "string", "format": "binary"},
							"dir":  map[string]any{"type": "string"},
						},
					},
				},
			},
		}
	default:
		return nil
	}
}

func mediaResponseSchemas(route routeinfo.RouteInfo) map[string]any {
	switch route.Method + " " + route.Path {
	case "POST /v1/media/upload", "GET /v1/media/url":
		return authResponses("#/components/schemas/SuccessResponse")
	case "DELETE /v1/media":
		return authResponses("#/components/schemas/SuccessResponse")
	default:
		return nil
	}
}
