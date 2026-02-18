package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/rQxwX3/chirpy/internal/database"
	"log"
	"net/http"
	"strings"
	"time"
)

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w,
		`<html>
		  <body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		  </body>
		</html>`, cfg.fileserverHits.Load(),
	)
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	cfg.handlerMetrics(w, r)
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func filterProfane(chirpBody string) string {
	profaneWords := [3]string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(chirpBody, " ")
	newWords := []string{}

	for _, word := range words {
		for _, profaneWord := range profaneWords {
			if strings.ToLower(word) == profaneWord {
				word = "****"
				break
			}
		}

		newWords = append(newWords, word)
	}

	return strings.Join(newWords, " ")
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	chirpStruct := chirp{}
	err := decoder.Decode(&chirpStruct)

	type resp struct {
		Valid       bool   `json:"valid"`
		Error       error  `json:"error"`
		CleanedBody string `json:"cleaned_body"`
	}

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(500)

		respBody := resp{
			Error: err,
		}

		data, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			return
		}

		w.Write(data)
		return
	}

	if len(chirpStruct.Body) > 140 {
		w.WriteHeader(400)

		respBody := resp{Valid: false}

		data, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			return
		}

		w.Write(data)
		return
	}

	w.WriteHeader(200)

	respBody := resp{CleanedBody: filterProfane(chirpStruct.Body)}
	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Write(data)
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)

	var reqStruct = req{}
	err := decoder.Decode(&reqStruct)
	if err != nil {
		log.Printf("Error decoding JSON: %s", err)
		return
	}

	type res struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     reqStruct.Email,
	})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		return
	}

	data, err := json.Marshal(user)
	if err != nil {
		log.Printf("Error mashalling JSON: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}
