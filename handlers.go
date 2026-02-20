package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/rQxwX3/chirpy/internal/auth"
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
		w.WriteHeader(500)
		log.Printf("Error reseting users table %s", err)
		return
	}
	log.Printf("Database reset")
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
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error obtaining JWT from headers: %s", err)
		return
	}

	userUUID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(401)
		log.Printf("Error validating JWT: %s", err)
		return
	}

	type req struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	reqStruct := req{}

	err = decoder.Decode(&reqStruct)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error decoding JSON: %s", err)
		return
	}

	if ok := validateChirp(&reqStruct.Body); !ok {
		w.WriteHeader(400)
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
		UserID:    userUUID,
	})
	if err != nil {
		w.WriteHeader(500)
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
		w.WriteHeader(500)
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)

	reqStruct := req{}
	err := decoder.Decode(&reqStruct)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error decoding JSON: %s", err)
		return
	}

	type res struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	hash, err := auth.HashPassword(reqStruct.Password)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error hashing the password: %s", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Email:          reqStruct.Email,
		HashedPassword: hash,
	})
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error creating user: %s", err)
		return
	}

	resBody := res{user.ID, user.CreatedAt, user.UpdatedAt, user.Email}
	data, err := json.Marshal(resBody)
	if err != nil {
		w.WriteHeader(500)
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
		w.WriteHeader(500)
		log.Printf("Error querying database for chirps")
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
		w.WriteHeader(500)
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error parsing UUID path value %s", err)
		return
	}

	chirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(404)
		log.Printf("Error querying database for chirp")
		return
	}

	type res struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	resBody := res{
		chirp.ID,
		chirp.CreatedAt,
		chirp.UpdatedAt,
		chirp.Body,
		chirp.UserID,
	}

	data, err := json.Marshal(resBody)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)

	reqStruct := req{}
	err := decoder.Decode(&reqStruct)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error decoding JSON: %s", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), reqStruct.Email)
	if err != nil {
		w.WriteHeader(404)
		log.Printf("Error querying database for user: %s", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Duration(3600)*time.Second)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error creating a JWT: %s", err)
		return
	}

	refreshTokenValue, err := auth.MakeRefreshToken()
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error creating a refresh token: %s", err)
		return
	}

	currentTime := time.Now().UTC()

	_, err = cfg.db.CreateRefreshToken(r.Context(),
		database.CreateRefreshTokenParams{
			Token:     refreshTokenValue,
			CreatedAt: currentTime,
			ExpiresAt: currentTime.Add(time.Hour * 24 * 60),
			UserID:    user.ID,
		})
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error inserting refresh token to the database: %s", err)
		return
	}

	ok, err := auth.CheckPasswordHash(reqStruct.Password, user.HashedPassword)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error checking password hash: %s", err)
		return
	}

	if !ok {
		w.WriteHeader(401)
		w.Write([]byte("Incorrect email or password"))
		return
	}

	type res struct {
		Id           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	resStruct := res{
		user.ID, user.CreatedAt, user.UpdatedAt, user.Email, token, refreshTokenValue,
	}

	data, err := json.Marshal(resStruct)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refreshTokenValue, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error getting refresh token from headers: %s", err)
		return
	}

	refreshToken, err := cfg.db.GetRefreshTokenByValue(r.Context(), refreshTokenValue)
	if err != nil {
		w.WriteHeader(401)
		log.Printf("Error retrieving refresh token from database: %s", err)
		return
	}

	if (time.Now().After(refreshToken.ExpiresAt) ||
		refreshToken.RevokedAt != sql.NullTime{}) {
		w.WriteHeader(401)
		return
	}

	token, err := auth.MakeJWT(refreshToken.UserID, cfg.jwtSecret,
		time.Duration(3600)*time.Second,
	)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error creating a JWT: %s", err)
		return
	}

	type res struct {
		Token string `json:"token"`
	}

	resStruct := res{token}

	data, err := json.Marshal(resStruct)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshTokenValue, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error getting refresh token from headers: %s", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), refreshTokenValue)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error revoking refresh token: %s", err)
		return
	}

	w.WriteHeader(204)
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Token    string `json:"token"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	reqStruct := req{}

	err := decoder.Decode(&reqStruct)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error decoding JSON: %s", err)
		return
	}

	userUUID, err := auth.ValidateJWT(reqStruct.Token, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	hash, err := auth.HashPassword(reqStruct.Password)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error hashing password: %s", err)
		return
	}

	user, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userUUID,
		Email:          reqStruct.Email,
		HashedPassword: hash,
	})
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error updating database: %s", err)
		return
	}

	type res struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	resStruct := res{user.ID, user.CreatedAt, user.UpdatedAt, user.Email}
	data, err := json.Marshal(resStruct)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}
