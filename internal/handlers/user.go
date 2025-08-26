package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

		//set to redis
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
