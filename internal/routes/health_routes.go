package routes

import (
	"net/http"

	"gotemp/internal/handlers"
)

func SetupHealthRoutes(mux *http.ServeMux, handler *handlers.Handler) {
	mux.HandleFunc("/health", handler.HealthHandler())
}
