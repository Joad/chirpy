package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/Joad/chirpy/internal/auth"
	"github.com/go-chi/chi/v5"
)

type Chirp struct {
	Id       int    `json:"id"`
	AuthorId int    `json:"author_id"`
	Body     string `json:"body"`
}

func (cfg *apiConfig) postChirp(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Body string `json:"body"`
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token required")
		return
	}
	subject, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error validating token")
		return
	}
	id, err := strconv.Atoi(subject)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid id")
		return
	}
	decoder := json.NewDecoder(r.Body)
	toValidate := params{}
	err = decoder.Decode(&toValidate)
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
	chirp, err := cfg.db.CreateChirp(replaceBadWords(toValidate.Body, badwords), id)
	if err != nil {
		log.Fatalln("Error creating chirp: ", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	respondWithJSON(w, 201, Chirp{
		Id:       chirp.Id,
		AuthorId: chirp.AuthorId,
		Body:     chirp.Body,
	})
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetChirps()

	authorId := -1
	authorIdParam := r.URL.Query().Get("author_id")
	if authorIdParam != "" {
		authorId, err = strconv.Atoi(authorIdParam)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error with author id")
			return
		}
	}

	chirps := []Chirp{}
	for _, chirp := range dbChirps {
		if authorId != -1 && chirp.AuthorId != authorId {
			continue
		}
		chirps = append(chirps, Chirp{
			Id:       chirp.Id,
			AuthorId: chirp.AuthorId,
			Body:     chirp.Body,
		})
	}
	sortFunc := func(i, j int) bool {
		return chirps[i].Id < chirps[j].Id
	}
	if r.URL.Query().Get("sort") == "desc" {
		sortFunc = func(i, j int) bool {
			return chirps[i].Id > chirps[j].Id
		}
	}
	sort.Slice(chirps, sortFunc)
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

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	s := chi.URLParam(r, "chirpid")

	chirpid, err := strconv.Atoi(s)
	if err != nil {
		log.Println("Error converting chirpid, ", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token required")
		return
	}
	subject, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error validating token")
		return
	}
	userId, err := strconv.Atoi(subject)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid id")
		return
	}

	chirp, found, err := cfg.db.GetChirpById(chirpid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting chirp")
		return
	}
	if !found {
		respondWithError(w, http.StatusNotFound, "Chirp does not exist")
		return
	}

	if chirp.AuthorId != userId {
		respondWithError(w, http.StatusForbidden, "Not authorized")
		return
	}

	err = cfg.db.DeleteChirp(chirpid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting chirp")
		return
	}
	respondWithJSON(w, http.StatusOK, struct{}{})
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
