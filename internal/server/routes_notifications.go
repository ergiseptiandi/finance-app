package server

import (
	"finance-backend/internal/auth"
	"finance-backend/internal/notifications"
	authmiddleware "finance-backend/internal/server/middleware"

	"github.com/go-chi/chi/v5"
)

func registerNotificationsRoutes(r chi.Router, catalog routeCatalog, authService *auth.Service, notificationService *notifications.Service) {
	if authService == nil || notificationService == nil {
		return
	}

	notifications.RegisterRoutes(r, notifications.HandlerDependencies{
		NotificationService: notificationService,
		AuthMiddleware:      authmiddleware.NewAuth(authService),
	})

	for _, route := range notifications.Definitions() {
		catalog.Add(route)
	}
}
