// package middlewares
//
// import (
// 	"net/http"
// )
//
// // TokenAuthMiddleware validates JWT token and checks blacklist in Redis
// func (h *Handler) TokenAuthMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		tokenString := extractTokenFromHeader(r)
// 		if tokenString == "" {
// 			http.Error(w, "Missing token", http.StatusUnauthorized)
// 			return
// 		}
//
// 		// Check if token is blacklisted
// 		blacklisted, err := h.Redis.Get(r.Context(), tokenString).Result()
// 		if err == nil && blacklisted == "blacklisted" {
// 			http.Error(w, "Token revoked", http.StatusUnauthorized)
// 			return
// 		}
//
// 		// You can optionally parse and validate the token here
// 		// and set user claims into context for downstream handlers
// 		// For brevity, skipping that step here, but you can add it
//
// 		next.ServeHTTP(w, r)
// 	})

// }

package middlewares
