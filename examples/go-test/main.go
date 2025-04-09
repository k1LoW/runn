package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func main() {
	http.Handle("/healthz", HealthCheck{})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           nil,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

type HealthCheck struct{}

func (HealthCheck) ServeHTTP(w http.ResponseWriter, r *http.Request) { //nostyle:recvtype
	resp := map[string]string{
		"message": "ok",
	}
	jsonBody, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(jsonBody); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
