package main

import (
	"encoding/json"
	"net/http"

	"github.com/Joad/chirpy/internal/auth"
)

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
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

	user, err := cfg.db.GetUserByEmail(params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Not allowed")
		return
	}

	if err = auth.CheckPassword(user.Password, params.Password); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Not allowed")
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Id:    user.Id,
		Email: user.Email,
	})
}
