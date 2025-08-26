package routes

import (
	"net/http"

	"gotemp/internal/handlers"
	"gotemp/internal/middlewares"
)

func SetupUserRoutes(mux *http.ServeMux, handler *handlers.Handler) {
	userMux := http.NewServeMux()

	// Define user routes with method-based routing
	userMux.HandleFunc("POST /register", handler.CreateUserHandler())
	userMux.HandleFunc("POST /login", handler.LoginUserHandler())

	userMux.Handle("POST /session/logout", middlewares.AuthMiddleware(http.HandlerFunc(handler.LogOutHandler())))
	userMux.Handle("GET /profile", middlewares.AuthMiddleware(http.HandlerFunc(handler.UserProfile())))

	mux.Handle("/users/", http.StripPrefix("/users", userMux))
}
