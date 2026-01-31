package main

import (
	"encoding/json"
	"log"
	"net/http"
)

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
