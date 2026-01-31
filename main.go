package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/devons3/chirpy/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error opening database: %s", err)
	}

	apiCfg := apiConfig{}
	apiCfg.dbQueries = database.New(db)

	servMux := http.NewServeMux()
	servMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	servMux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	servMux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	servMux.HandleFunc("GET /api/healthz", handlerReady)
	servMux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	var server http.Server

	server.Handler = servMux
	server.Addr = ":8080"

	server.ListenAndServe()

}

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
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

func cleanChirp(chirpOrig string) (bool, string) {
	words := strings.Split(chirpOrig, " ")
	containsProfanity := false
	for i := 0; i < len(words); i++ {
		word := strings.ToLower(words[i])
		switch word {
		case "kerfuffle", "sharbert", "fornax":
			containsProfanity = true
			words[i] = "****"
		}
	}
	cleanedChirp := strings.Join(words, " ")
	return containsProfanity, cleanedChirp
}

func respondWithError(respWriter http.ResponseWriter, code int, msg string) {

	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respBody := errorResponse{
		Error: msg,
	}

	respondWithJSON(respWriter, code, respBody)
}

func respondWithJSON(respWriter http.ResponseWriter, code int, payload interface{}) {
	respWriter.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		respWriter.WriteHeader(500)
		return
	}
	respWriter.WriteHeader(code)
	respWriter.Write(data)
}
