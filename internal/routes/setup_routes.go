package routes

import (
	"net/http"

	"gotemp/internal/errorrhandler"
	"gotemp/internal/handlers"
)

func SetupRoutes(mux *http.ServeMux, handler *handlers.Handler) {
	SetupUserRoutes(mux, handler)
	SetupHealthRoutes(mux, handler)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		errorrhandler.RespondWithNotFound(w)
	})

}
