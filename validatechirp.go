package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type chirpBody struct {
		Body string `json:"body"`
	}
	type chirpCleaned struct {
		Cleaned string `json:"cleaned_body"`
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

	badwords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	validResp := chirpCleaned{
		Cleaned: replaceBadWords(toValidate.Body, badwords),
	}
	respondWithJSON(w, 200, validResp)
}

func replaceBadWords(chirp string, badwords map[string]bool) string {
	words := strings.Split(chirp, " ")
	cleanedWords := make([]string, 0, len(words))
	for _, word := range words {
		if _, found := badwords[strings.ToLower(word)]; found {
			word = "****"
		}
		cleanedWords = append(cleanedWords, word)
	}
	return strings.Join(cleanedWords, " ")
}
