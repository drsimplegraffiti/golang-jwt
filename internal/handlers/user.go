package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"gotemp/internal/auth"
	"gotemp/internal/dtos/request"
	"gotemp/internal/errorrhandler"
	"gotemp/internal/middlewares"
	"gotemp/internal/store"
	"gotemp/internal/successresponse"
	"gotemp/internal/utils"
	"gotemp/internal/validation"
)

// Helper: extract token from Authorization header
func extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

// cleanUserSessions deletes all Redis session keys for the given userID
func (h *Handler) cleanUserSessions(userID string) error {
	pattern := fmt.Sprintf("session:%s:*", userID)
	ctx := context.Background()

	iter := h.Redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		err := h.Redis.Del(ctx, iter.Val()).Err()
		if err != nil {
			// Log error and continue deleting other keys
			fmt.Printf("failed to delete session key %s: %v\n", iter.Val(), err)
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return nil
}

func (h *Handler) LogOutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract JWT claims from context (set by AuthMiddleware)
		claims, ok := r.Context().Value(middlewares.UserClaimsKey).(*auth.Claims)
		if !ok {
			errorrhandler.RespondWithError(w, http.StatusBadRequest, "Please login to continue")
			return
		}

		// Extract token string from Authorization header
		tokenString := extractTokenFromHeader(r)
		if tokenString == "" {
			errorrhandler.RespondWithError(w, http.StatusUnauthorized, "Missing token")
			return
		}

		// Convert ExpiresAt to time.Time
		expirationTime := time.Unix(claims.ExpiresAt, 0)
		now := time.Now()
		ttl := expirationTime.Sub(now)
		if ttl <= 0 {
			ttl = 5 * time.Minute // fallback TTL
		}

		// Blacklist the token in Redis
		err := h.Redis.Set(r.Context(), tokenString, "blacklisted", ttl).Err()
		if err != nil {
			errorrhandler.RespondWithError(w, http.StatusInternalServerError, "Failed to blacklist token")
			return
		}

		// Clean user sessions in Redis
		userIDStr := fmt.Sprintf("%d", claims.UserID)
		if err := h.cleanUserSessions(userIDStr); err != nil {
			fmt.Printf("Error cleaning sessions for user %s: %v\n", userIDStr, err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Logged out successfully"))
	}
}

// profile
func (h *Handler) UserProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(middlewares.UserClaimsKey).(*auth.Claims)
		if !ok {
			errorrhandler.RespondWithError(w, http.StatusBadRequest, "please login to continue")
			return
		}
		userID := claims.UserID

		// Check Redis cache first
		cacheKey := fmt.Sprintf("user:%d", userID)
		if cached, err := h.Redis.Get(r.Context(), cacheKey).Result(); err == nil {
			// Cached value exists
			var user store.User
			if err := json.Unmarshal([]byte(cached), &user); err == nil {
				successresponse.RespondWithSuccess(w, http.StatusOK, "success (from cache)", user)
				return
			}
		}

		// Fallback to DB
		user, err := h.Queries.GetUser(r.Context(), int32(userID))
		if err != nil {
			errorrhandler.RespondWithError(w, http.StatusNotFound, "user not found")
			return
		}

		// Save in Redis for next time
		userJSON, _ := json.Marshal(user)
		h.Redis.Set(r.Context(), cacheKey, userJSON, 5*time.Minute)

		successresponse.RespondWithSuccess(w, http.StatusOK, "success", user)
	}
}

// login a user
func (h *Handler) LoginUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req request.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorrhandler.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// Validate the request
		if err := validation.Validate(&req); err != nil {
			errorrhandler.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// fetch the user from the db using the store queries
		user, err := h.Queries.GetUserByUsernameOrEmail(ctx, req.Username)
		if err != nil {
			errorrhandler.RespondWithError(w, http.StatusUnauthorized, "invalid credential")
			return
		}

		if !utils.ComparePassword(user.Password, req.Password) {
			errorrhandler.RespondWithError(w, http.StatusUnauthorized, "invalid credential")
			return
		}

		jwtKey := []byte(os.Getenv("JWT_SECRET_KEY"))
		token, err := auth.GenerateJWT(int64(user.ID), user.Username, jwtKey)
		if err != nil {
			errorrhandler.RespondWithError(w, http.StatusInternalServerError, "Error generating a token")
			return
		}

		successresponse.RespondWithSuccess(w, http.StatusOK, "Login successful", map[string]string{
			"token": token,
		})
	}
}

// CreateUserHandler with Transaction
func (h *Handler) CreateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req request.CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorrhandler.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// Validate the request
		if err := validation.Validate(&req); err != nil {
			errorrhandler.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Start a transaction
		tx, err := h.DB.BeginTx(ctx, nil)
		if err != nil {
			errorrhandler.RespondWithError(w, http.StatusInternalServerError, "Failed to start transaction")
			return
		}
		defer tx.Rollback() // Ensures rollback if any error occurs

		// Create a new Queries instance bound to the transaction
		qtx := store.New(tx)

		// Check if the username already exists
		_, err = qtx.GetUserByUsernameOrEmail(ctx, req.Username)
		if err == nil {
			errorrhandler.RespondWithError(w, http.StatusConflict, "Username already taken")
			return
		}

		// Check if the email already exists
		_, err = qtx.GetUserByUsernameOrEmail(ctx, req.Email)
		if err == nil {
			errorrhandler.RespondWithError(w, http.StatusConflict, "Email already taken")
			return
		}

		// Hash password
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			errorrhandler.RespondWithError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		// Create user within the transaction
		_, err = qtx.CreateUser(ctx, store.CreateUserParams{
			Username: req.Username,
			Email:    req.Email,
			Password: string(hashedPassword),
		})
		if err != nil {
			errorrhandler.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		// Commit the transaction if all operations succeed
		if err := tx.Commit(); err != nil {
			errorrhandler.RespondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
			return
		}

		successresponse.RespondWithSuccess(w, http.StatusCreated, "User created successfully", nil)
	}
}
