package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}

	type resp struct {
		Valid bool `json:"valid"`
	}

	type errorResp struct {
		Error error `json:"error"`
	}

	decoder := json.NewDecoder(r.Body)
	chirpStruct := chirp{}
	err := decoder.Decode(&chirpStruct)

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(500)

		respBody := errorResp{
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

	respBody := resp{Valid: true}

	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Write(data)
}
