package main

import (
	// "github.com/google/uuid"
	"database/sql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rQxwX3/chirpy/internal/database"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func getDBQueries(dbURL string) (*database.Queries, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	return database.New(db), nil
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	queries, err := getDBQueries(dbURL)
	if err != nil {
		log.Fatalf("Error connecting to the database: %s", err)
	}

	cfg := apiConfig{
		db:       queries,
		platform: os.Getenv("PLATFORM"),
	}

	mux := http.NewServeMux()

	fsHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", cfg.middlewareMetricsInc(fsHandler))

	mux.HandleFunc("GET /api/healthz", handlerHealth)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)
	mux.HandleFunc("POST /api/users", cfg.handlerCreateUser)
	mux.HandleFunc("POST /api/chirps", cfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", cfg.handlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerGetChirpByID)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Server error:", err)
	}
}
