package main

import (
	"encoding/json"
	"net/http"
)

func main() {
	http.Handle("/healthz", HealthCheck{})
	http.ListenAndServe(":8080", nil)
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
	w.Write(jsonBody)
}
