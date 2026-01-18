package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Sergyrm/chirpy/internal/auth"
	"github.com/Sergyrm/chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerUserCreate(w http.ResponseWriter, r *http.Request) {
	params, err := decodeUser(r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	user, err := cfg.createUser(r.Context(), params.Email, hashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, user)
}

func (cfg *apiConfig) createUser(context context.Context, email string, hashedPassword string) (*User, error) {
	user, err := cfg.db.CreateUser(context, database.CreateUserParams{
		Email:          email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}, nil
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type response struct {
		User
		Token 		 string `json:"token"`
		RefreshToken string `json:"refresh_token,omitempty"`
	}
	params, err := decodeUser(r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.ExpiresInSeconds <= 0 || params.ExpiresInSeconds > 3600 {
		params.ExpiresInSeconds = 3600
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password", nil)
		return
	}

	if ok, err := auth.CheckPasswordHash(params.Password, user.HashedPassword); !ok {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.tokenSecret, time.Duration(params.ExpiresInSeconds)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create token", err)
		return
	}
	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token: token,
	})
}

func decodeUser(r *http.Request) (struct {
										Password string `json:"password"`
										Email    string `json:"email"`
										ExpiresInSeconds int64  `json:"expires_in_seconds,omitempty"`
									}, error) {
	type parameters struct {
		Password 		 string `json:"password"`
		Email    		 string `json:"email"`
		ExpiresInSeconds int64  `json:"expires_in_seconds,omitempty"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		return parameters{}, err
	}
	return params, nil
}