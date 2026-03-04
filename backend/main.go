package main

import (
	"fmt"
	"log/slog"
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
			slog.Error("Failed to connect to database", "error", err)
			os.Exit(1)
		}
		defer conn.Close()

		if err := db.RunMigrations(conn); err != nil {
			slog.Error("Failed to run migrations", "error", err)
			os.Exit(1)
		}

		repo = db.NewRepository(conn)
		slog.Info("Database initialized successfully")

		// Auth service requires a database
		adapter := auth.NewDBAdapter(repo)
		authSvc = auth.NewAuthService(adapter)
		slog.Info("Auth service initialized")
	} else {
		slog.Warn("DATABASE_URL not set — running without database persistence or auth")
	}

	hub := room.NewHub(repo)
	go hub.Run()
	hub.StartCleanup()

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
		slog.Info("Serving static files from ./static")
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
	slog.Info("Backend running", "addr", addr)
	if err := http.ListenAndServe(addr, api.RequestIDMiddleware(corsMiddleware(mux))); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
