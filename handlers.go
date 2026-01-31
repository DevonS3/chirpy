package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

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

func handlerReady(resp_writeer http.ResponseWriter, _ *http.Request) {
	resp_writeer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp_writeer.WriteHeader(200)
	resp_writeer.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerReset(respWriter http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits.Store(0)
	respWriter.WriteHeader(http.StatusOK)
	respWriter.Write([]byte("Hits reset to 0"))
}

func handlerValidateChirp(respWriter http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type parameters struct {
		Body string `json:"body"`
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
