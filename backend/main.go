package main

import (
	"log"
	"net/http"
	"sabacc/api"
	"sabacc/room"
)

func main() {
	hub := room.NewHub()
	go hub.Run()

	handler := api.NewHandler(hub)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// CORS middleware for local dev
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	log.Println("Backend running on :8080")
	if err := http.ListenAndServe(":8080", corsMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}
