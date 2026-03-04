package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sabacc/api"
	"sabacc/auth"
	"sabacc/db"
	"sabacc/room"
)

func main() {
	// --- Optional database initialization ---
	var repo *db.Repository
	var authSvc *auth.AuthService
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		conn, err := db.Connect(databaseURL)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer conn.Close()

		if err := db.RunMigrations(conn); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}

		repo = db.NewRepository(conn)
		log.Println("Database initialized successfully")

		// Auth service requires a database
		adapter := auth.NewDBAdapter(repo)
		authSvc = auth.NewAuthService(adapter)
		log.Println("Auth service initialized")
	} else {
		log.Println("WARNING: DATABASE_URL not set — running without database persistence or auth")
	}

	hub := room.NewHub(repo)
	go hub.Run()

	handler := api.NewHandler(hub, repo, authSvc)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Serve static frontend files in production (built assets in ./static)
	if info, err := os.Stat("./static"); err == nil && info.IsDir() {
		staticFS := http.FileServer(http.Dir("./static"))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Try to serve the file directly; fall back to index.html for SPA routing
			path := "./static" + r.URL.Path
			if _, err := os.Stat(path); os.IsNotExist(err) {
				http.ServeFile(w, r, "./static/index.html")
				return
			}
			staticFS.ServeHTTP(w, r)
		})
		log.Println("Serving static files from ./static")
	}

	// CORS middleware for local dev
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Backend running on %s", addr)
	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}
