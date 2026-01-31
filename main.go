package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/devons3/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error opening database: %s", err)
	}

	apiCfg := apiConfig{}
	apiCfg.dbQueries = database.New(db)

	servMux := http.NewServeMux()
	servMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("static")))))
	servMux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	servMux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	servMux.HandleFunc("GET /api/healthz", handlerReady)
	servMux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	servMux.HandleFunc("POST /api/chirps", handlerChirps)
	servMux.HandleFunc("POST /api/users", apiCfg.handlerUsers)

	var server http.Server

	server.Handler = servMux
	server.Addr = ":8080"

	server.ListenAndServe()

}

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
}
