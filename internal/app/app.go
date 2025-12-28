package app

import (
	"context"
	"database/sql"
	"gostart/internal/auth"
	"gostart/internal/user"
	"net/http"
	"time"
)

type App struct {
	server *http.Server
}

func New(dsn string) (*App, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	userRepo := user.NewPostgres(db)
	userService := user.NewService(userRepo)

	mux := http.NewServeMux()
	mux.Handle("/login", user.LoginHandler(userService))
	mux.Handle("/register", user.RegisterHandler(userService))
	mux.Handle("/user", auth.Middleware(user.GetHandler(userService)))

	handler := cors(mux)

	server := &http.Server{
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	return &App{server: server}, nil
}

func (a *App) Run(addr string) error {
	a.server.Addr = addr
	return a.server.ListenAndServe()
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
	})
}
