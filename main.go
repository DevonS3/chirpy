package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
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
