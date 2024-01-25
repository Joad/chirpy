package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/Joad/chirpy/internal/database"
	"github.com/go-chi/chi/v5"
)

func main() {
	const root = "."
	const port = "8080"
	const path = "database.json"

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		err := os.Remove(path)
		if err != nil && !os.IsNotExist(err) {
			log.Fatal("Error removing database", err)
			return
		}
	}
	db, err := database.NewDB(path)
	if err != nil {
		log.Fatal("Error creating database: ", err)
		return
	}
	apiCfg := &apiConfig{
		db: db,
	}

	r := chi.NewRouter()

	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(root))))
	r.Handle("/app/*", fsHandler)
	r.Handle("/app", fsHandler)

	rApi := chi.NewRouter()
	rApi.Get("/metrics", apiCfg.writeMetrics())
	rApi.Handle("/reset", apiCfg.reset())
	rApi.Get("/healthz", healthz)

	rApi.Post("/chirps", apiCfg.postChirp)
	rApi.Get("/chirps", apiCfg.getChirps)
	rApi.Get("/chirps/{chirpid}", apiCfg.getChirp)

	rApi.Post("/users", apiCfg.postUsers)

	rApi.Post("/login", apiCfg.login)

	rAdmin := chi.NewRouter()
	rAdmin.Get("/metrics", apiCfg.htmlMetrics())

	r.Mount("/api", rApi)
	r.Mount("/admin", rAdmin)

	mux := middlewareCors(r)

	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", root, port)
	log.Fatal(server.ListenAndServe())
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	type chirpError struct {
		Error string `json:"error"`
	}
	errorResp := chirpError{
		Error: msg,
	}
	dat, err := json.Marshal(errorResp)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %s", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %s", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(dat)
}
