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

func extractTokenFromHader(r *http.Request) string {
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

// cleanUserSession
func (h *Handler) cleanUserSession(userID string) error {
	// session:123:*
	pattern := fmt.Sprintf("session:%s:*", userID)

	// Background context for redis
	ctx := context.Background()

	// scan to iterate over all the keys matching the patter declared
	iter := h.Redis.Scan(ctx, 0, pattern, 0).Iterator()

	// loop through each key from redis
	for iter.Next(ctx) {

		// delete the key from redihj
		err := h.Redis.Del(ctx, iter.Val()).Err()
		if err != nil {
			fmt.Printf("failed to delete session")
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	return nil
}

// logout handler
func (h *Handler) LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract the jwt claims from the context
		claims, ok := r.Context().Value(middlewares.UserClaimsKey).(*auth.Claims)
		if !ok {
			errorrhandler.RespondWithError(w, http.StatusBadRequest, "Please login to continue")
			return
		}

		// extract the token from the auth header
		tokenString := extractTokenFromHader(r)
		if tokenString == "" {
			errorrhandler.RespondWithError(w, http.StatusUnauthorized, "Missing token")
			return
		}

		// convert expireAt to time.Time
		expirationTime := time.Unix(claims.ExpiresAt, 0)
		now := time.Now()
		ttl := expirationTime.Sub(now)
		if ttl <= 0 {
			ttl = 5 * time.Minute // fallbask ttl
		}

		// Blacklist the token in redis
		err := h.Redis.Set(r.Context(), tokenString, "blacklisted", ttl).Err()
		if err != nil {
			errorrhandler.RespondWithError(w, http.StatusInternalServerError, "failed to blacklist token")
			return
		}

		// clean user session in Redis
		userIDStr := fmt.Sprintf("%d", claims.UserID)
		if err := h.cleanUserSession(userIDStr); err != nil {
			fmt.Printf("Error cleaning session for %s: %v\n", userIDStr, err)
		}

		successresponse.RespondWithSuccess(w, http.StatusOK, "Logged out successfully", true)
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

		// check the redis first
		cacheKey := fmt.Sprintf("user:%d", userID)
		if cached, err := h.Redis.Get(r.Context(), cacheKey).Result(); err == nil {
			fmt.Println("Inside the redisc func")
			var user store.User
			if err := json.Unmarshal([]byte(cached), &user); err == nil {
				successresponse.RespondWithSuccess(w, http.StatusOK, "success (from cache/redis)", user)
				return
			}
		}

		// Fallback to DB
		user, err := h.Queries.GetUser(r.Context(), int32(userID))
		if err != nil {
			errorrhandler.RespondWithError(w, http.StatusNotFound, "user not found")
			return
		}

		// set to redis
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
