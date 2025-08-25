package middlewares

import (
	"context"
	"net/http"
	"os"
	"strings"

	"gotemp/internal/auth"
	"gotemp/internal/errorrhandler"

	"github.com/dgrijalva/jwt-go"
)

// creates a custom type for context key to avoid collision
type contextKey string

// constant used in storing our user claims
const UserClaimsKey contextKey = "claims"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// retrieves the Authorization header from the request (postman/web/mobile)
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errorrhandler.RespondWithError(w, http.StatusUnauthorized, "No token provided")
			return
		}

		// strips the Bearer from the Bearer token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &auth.Claims{}

		// Parse the token and also validate it
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// we provide our key from the environment variable and validate it against the token from the request
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})
		// handle the validation error
		if err != nil {
			// handling likely tampered token
			if err == jwt.ErrSignatureInvalid {
				errorrhandler.RespondWithError(w, http.StatusBadRequest, "invalid token signature")
				return
			}

			// handles any other parsing error e.g expired, malformed etc
			errorrhandler.RespondWithError(w, http.StatusBadRequest, "invalid token")
			return
		}

		// if token is valid, store the claims in the request context
		if token.Valid {
			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			r = r.WithContext(ctx) // replace request context with the new one
			next.ServeHTTP(w, r)   // calls the next handler, with the updated request
		} else {
			errorrhandler.RespondWithError(w, http.StatusUnauthorized, "invalid token")
		}
	})
}
