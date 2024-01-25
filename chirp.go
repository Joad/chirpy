package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (cfg *apiConfig) postChirp(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	toValidate := params{}
	err := decoder.Decode(&toValidate)
	if err != nil {
		log.Printf("Error decoding chirp body: %s\n", err)
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
	chirp, err := cfg.db.CreateChirp(replaceBadWords(toValidate.Body, badwords))
	if err != nil {
		log.Fatalln("Error creating chirp: ", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	respondWithJSON(w, 201, chirp)
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps()
	if err != nil {
		log.Fatalln("Error getting chirps, ", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	s := chi.URLParam(r, "chirpid")

	chirpid, err := strconv.Atoi(s)
	if err != nil {
		log.Println("Error converting chirpid, ", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	chirp, found, err := cfg.db.GetChirpById(chirpid)
	if err != nil {
		log.Fatalln("Error retrieving chirp, ", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if !found {
		respondWithError(w, 404, "Chirp not found")
		return
	}

	respondWithJSON(w, 200, chirp)
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
