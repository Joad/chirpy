package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	const root = "."
	const port = "8080"
	apiCfg := &apiConfig{}

	r := chi.NewRouter()

	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(root))))
	r.Handle("/app/*", fsHandler)
	r.Handle("/app", fsHandler)

	rApi := chi.NewRouter()
	rApi.Get("/metrics", apiCfg.writeMetrics())
	rApi.Handle("/reset", apiCfg.reset())
	rApi.Get("/healthz", healthz)

	r.Mount("/api", rApi)

	mux := middlewareCors(r)

	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", root, port)
	log.Fatal(server.ListenAndServe())
}
