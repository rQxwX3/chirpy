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
	if cfg.platform != "dev" {
		w.WriteHeader(403)
		return
	}

	cfg.fileserverHits.Store(0)

	err := cfg.db.DeleteAll(r.Context())
	if err != nil {
		log.Printf("Error reseting users table %s", err)
		return
	}
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

func validateChirp(chirpBody *string) bool {
	if len(*chirpBody) > 140 {
		return false
	}

	*chirpBody = filterProfane(*chirpBody)

	return true
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	reqStruct := req{}

	err := decoder.Decode(&reqStruct)
	if err != nil {
		log.Printf("Error decoding JSON: %s", err)
		return
	}

	if ok := validateChirp(&reqStruct.Body); !ok {
		log.Printf("Error creating chirp: body exceeds max length")
		return
	}

	type res struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      reqStruct.Body,
		UserID:    reqStruct.UserID,
	})
	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		return
	}

	resBody := res{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	data, err := json.Marshal(resBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
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

	resBody := res{user.ID, user.CreatedAt, user.UpdatedAt, user.Email}
	data, err := json.Marshal(resBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Error quering database for chips")
		return
	}

	type res struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	resBody := []res{}
	for _, chirp := range chirps {
		resBody = append(resBody, res{
			chirp.ID,
			chirp.CreatedAt,
			chirp.UpdatedAt,
			chirp.Body,
			chirp.UserID,
		})
	}

	data, err := json.Marshal(resBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
