package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Joad/chirpy/internal/auth"
)

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		Id           int    `json:"id"`
		Email        string `json:"email"`
		IsChirpyRed  bool   `json:"is_chirpy_red"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode params")
		return
	}

	user, err := cfg.db.GetUserByEmail(params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Not allowed")
		return
	}

	if err = auth.CheckPassword(user.Password, params.Password); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Not allowed")
		return
	}

	now := time.Now().UTC()
	expiresIn := 1 * time.Hour
	expiresAt := now.Add(expiresIn)

	tokenString, err := auth.MakeJWT(user.Id, now, expiresAt,
		auth.AccessType, cfg.jwtSecret)
	if err != nil {
		log.Println("Error signing token, ", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	expiresIn = 60 * 24 * time.Hour
	expiresAt = now.Add(expiresIn)
	refreshTokenString, err := auth.MakeJWT(user.Id, now, expiresAt,
		auth.RefreshType, cfg.jwtSecret)
	if err != nil {
		log.Println("Error signing token, ", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Id:           user.Id,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        tokenString,
		RefreshToken: refreshTokenString,
	})
}

func (cfg *apiConfig) refresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Bearer required")
		return
	}

	revoked, err := cfg.db.IsTokenRevoked(token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error with db")
		return
	}

	if revoked {
		respondWithError(w, http.StatusUnauthorized, "Token revoked")
		return
	}

	tokenString, err := auth.RefreshToken(token, cfg.jwtSecret)
	if err != nil {
		log.Println("Error refreshing token, ", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: tokenString,
	})
}

func (cfg *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Bearer required")
		return
	}

	err = cfg.db.RevokeToken(token, time.Now().UTC())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error revoking token")
		return
	}
	respondWithJSON(w, http.StatusOK, struct{}{})
}
