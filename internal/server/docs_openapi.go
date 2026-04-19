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

		if parameters := parameterSchemas(route); parameters != nil {
			operation["parameters"] = parameters
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

func openAPIComponents() map[string]any {
	components := map[string]any{
		"securitySchemes": map[string]any{},
		"schemas":         map[string]any{},
	}

	mergeOpenAPIComponents(components, coreOpenAPIComponents())
	mergeOpenAPIComponents(components, authOpenAPIComponents())
	mergeOpenAPIComponents(components, transactionOpenAPIComponents())
	mergeOpenAPIComponents(components, categoryOpenAPIComponents())
	mergeOpenAPIComponents(components, salaryOpenAPIComponents())
	mergeOpenAPIComponents(components, debtOpenAPIComponents())
	mergeOpenAPIComponents(components, dashboardOpenAPIComponents())
	mergeOpenAPIComponents(components, reportsOpenAPIComponents())
	mergeOpenAPIComponents(components, alertsOpenAPIComponents())
	mergeOpenAPIComponents(components, notificationsOpenAPIComponents())
	mergeOpenAPIComponents(components, mediaOpenAPIComponents())

	return components
}

func mergeOpenAPIComponents(dst, src map[string]any) {
	for key, value := range src {
		switch typed := value.(type) {
		case map[string]any:
			target, _ := dst[key].(map[string]any)
			if target == nil {
				target = map[string]any{}
				dst[key] = target
			}

			for nestedKey, nestedValue := range typed {
				target[nestedKey] = nestedValue
			}
		default:
			dst[key] = value
		}
	}
}

func operationID(route routeinfo.RouteInfo) string {
	if id, ok := coreOperationID(route); ok {
		return id
	}
	if id, ok := authOperationID(route); ok {
		return id
	}
	return methodName(route.Method) + route.Path
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
	if requestBody := coreRequestBodySchema(route); requestBody != nil {
		return requestBody
	}
	if requestBody := authRequestBodySchema(route); requestBody != nil {
		return requestBody
	}
	if requestBody := transactionRequestBodySchema(route); requestBody != nil {
		return requestBody
	}
	if requestBody := categoryRequestBodySchema(route); requestBody != nil {
		return requestBody
	}
	if requestBody := salaryRequestBodySchema(route); requestBody != nil {
		return requestBody
	}
	if requestBody := debtRequestBodySchema(route); requestBody != nil {
		return requestBody
	}
	if requestBody := dashboardRequestBodySchema(route); requestBody != nil {
		return requestBody
	}
	if requestBody := reportsRequestBodySchema(route); requestBody != nil {
		return requestBody
	}
	if requestBody := alertsRequestBodySchema(route); requestBody != nil {
		return requestBody
	}
	if requestBody := notificationsRequestBodySchema(route); requestBody != nil {
		return requestBody
	}
	if requestBody := mediaRequestBodySchema(route); requestBody != nil {
		return requestBody
	}

	return nil
}

func responseSchemas(route routeinfo.RouteInfo) map[string]any {
	if responses := coreResponseSchemas(route); responses != nil {
		return responses
	}
	if responses := authResponseSchemas(route); responses != nil {
		return responses
	}
	if responses := transactionResponseSchemas(route); responses != nil {
		return responses
	}
	if responses := categoryResponseSchemas(route); responses != nil {
		return responses
	}
	if responses := salaryResponseSchemas(route); responses != nil {
		return responses
	}
	if responses := debtResponseSchemas(route); responses != nil {
		return responses
	}
	if responses := dashboardResponseSchemas(route); responses != nil {
		return responses
	}
	if responses := reportsResponseSchemas(route); responses != nil {
		return responses
	}
	if responses := alertsResponseSchemas(route); responses != nil {
		return responses
	}
	if responses := notificationsResponseSchemas(route); responses != nil {
		return responses
	}
	if responses := mediaResponseSchemas(route); responses != nil {
		return responses
	}

	return nil
}

func parameterSchemas(route routeinfo.RouteInfo) []map[string]any {
	if parameters := transactionParameterSchemas(route); parameters != nil {
		return parameters
	}

	return nil
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
