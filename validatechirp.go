package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type chirpBody struct {
		Body string `json:"body"`
	}
	type chirpValid struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	toValidate := chirpBody{}
	err := decoder.Decode(&toValidate)
	if err != nil {
		log.Printf("Error decoding chirp body: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if len(toValidate.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	validResp := chirpValid{
		Valid: true,
	}
	respondWithJSON(w, 200, validResp)
}
