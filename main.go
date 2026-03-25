package main

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/P8ace/RateLimiterGo/limiter"
)

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

func APIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	w.WriteHeader(http.StatusOK)
	message := Message{
		Status: "Success",
		Body:   "You have reached the API successfully",
	}
	err := json.NewEncoder(w).Encode(&message)
	if err != nil {
		slog.Error("Error while encoding the message", "Error: ", err.Error())
	}
}

func main() {
	mux := http.DefaultServeMux

	mux.Handle("/ping", limiter.RateLimiterFromX(APIHandler))
	slog.Info("Starting the http server", "Port", "8080")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		slog.Error("Error while starting the http server", "Error details:", err.Error())
	}
}
