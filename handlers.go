package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerMetrics(respWriter http.ResponseWriter, _ *http.Request) {
	body := fmt.Sprintf(`
		<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
		</html>
		`, cfg.fileserverHits.Load())
	respWriter.Header().Add("Content-Type", "text/html; charset=utf-8")
	respWriter.WriteHeader(http.StatusOK)
	respWriter.Write([]byte(body))
}

func handlerReady(respWriter http.ResponseWriter, _ *http.Request) {
	respWriter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	respWriter.WriteHeader(200)
	respWriter.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerReset(respWriter http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	err := cfg.dbQueries.DeleteAllUsers(r.Context())

	if err != nil {
		respondWithError(respWriter, 500, "Failed to delete users")
		return
	}

	respWriter.WriteHeader(http.StatusOK)
	respWriter.Write([]byte("Hits reset to 0"))
}

func (cfg *apiConfig) handlerUsers(respWriter http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		log.Printf("Error decoding body: %s", err)
		respWriter.WriteHeader(500)
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), params.Email)

	if err != nil {
		respondWithError(respWriter, 500, fmt.Sprint(err))
		return
	}

	respBody := UserResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(respWriter, 201, respBody)

}

func handlerChirps(respWriter http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type parameters struct {
		Body string `json:"body"`
		ID   string `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		log.Printf("Error decoding body: %s", err)
		respWriter.WriteHeader(500)
		return
	}

	type returnVals struct {
		Cleaned_Body string `json:"cleaned_body,omitempty"`
	}

	respBody := returnVals{}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(respWriter, 400, "Chirp is too long.")
		return
	}

	_, cleanedChirp := cleanChirp(params.Body)
	respBody.Cleaned_Body = cleanedChirp
	respondWithJSON(respWriter, 200, respBody)
}
