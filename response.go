package main

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Response functions from: https://github.com/krishbhanushali/go-rest-unit-testing/blob/master/api.go
// RespondWithError is called on an error to return info regarding error
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// Called for responses to encode and send json data
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	//encode payload to json
	response, _ := json.Marshal(payload)

	// set headers and write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		log.Errorf("w.Write() response failed: %v", err)
	}
}
