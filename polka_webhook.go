package main

import (
	"encoding/json"
	"net/http"

	"github.com/Joad/chirpy/internal/auth"
)

func (cfg *apiConfig) polkaWebhook(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserId int `json:"user_id"`
		} `json:"data"`
	}

	err := auth.ValidatePolkaKey(r.Header, cfg.polkaKey)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error with key")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode params")
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusOK, struct{}{})
		return
	}

	err = cfg.db.UpgradeUser(params.Data.UserId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}
	respondWithJSON(w, http.StatusOK, struct{}{})
}
