package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/Joad/chirpy/internal/auth"
)

func (cfg *apiConfig) postUsers(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		Id          int    `json:"id"`
		Email       string `json:"email"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Println("Error decoding params: ", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode params")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Println("Error hashing password: ", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password")
	}

	user, err := cfg.db.CreateUser(params.Email, hashedPassword)
	if err != nil {
		log.Println("Error creating user: ", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		Id:          user.Id,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode params")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Bearer required")
		return
	}

	subject, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		log.Println("Error validating jwt: ", err)
		respondWithError(w, http.StatusUnauthorized, "Error with token")
		return
	}

	id, err := strconv.Atoi(subject)
	if err != nil {
		log.Println("Error getting user id: ", err)
		respondWithError(w, http.StatusInternalServerError, "Error getting user id")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password")
	}

	user, err := cfg.db.UpdateUser(id, params.Email, hashedPassword)
	if err != nil {
		log.Println("Error updating user: ", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Id:    user.Id,
		Email: user.Email,
	})
}
