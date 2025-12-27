package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"gostart/internal/user"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Age      int64  `json:"age"`
	Password string `json:"password"`
}

var db *sql.DB

// в ПРОДЕ обязательно
// env-переменная
//
// длинный ключ
//
// не в коде
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type contextKey string

const userIDKey contextKey = "userID"

func getUserHandler(w http.ResponseWriter, r *http.Request, svc user.Service) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	u, err := svc.GetUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "DB error GetUser", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := int64(claims["user_id"].(float64))

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
func loginHandler(w http.ResponseWriter, r *http.Request, svc user.Service) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	u, err := svc.Login(r.Context(), creds.Username, creds.Password)
	if err != nil {
		http.Error(w, "DB error Login", http.StatusInternalServerError)
		return
	}
	if u == nil {
		http.Error(w, "Invalid credentials", http.StatusNotFound)
		return
	}
	claims := jwt.MapClaims{
		"user_id": u.ID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "token invalid", http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}
func registerHandler(w http.ResponseWriter, r *http.Request, svc user.Service) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Age      int64  `json:"age"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	u, err := svc.Register(r.Context(), creds.Username, creds.Password, creds.Age)
	if err != nil {
		if errors.Is(err, user.ErrUserAlreadyExists) {
			http.Error(w, "User already exists", http.StatusConflict) // 409
			return
		}
		http.Error(w, "DB error Register", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{
		"id": u.ID,
	})

}
func main() {
	var err error
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	userRepo := user.NewPostgres(db)
	service := user.NewService(userRepo)
	log.Println("Connected to postgres")

	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		loginHandler(w, r, service)
	})
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		registerHandler(w, r, service)
	})
	mux.Handle("/user",
		authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			getUserHandler(w, r, service)
		})),
	)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == http.MethodOptions {
			return
		}

		mux.ServeHTTP(w, r)
	})
	http.ListenAndServe(":8080", handler)
	//http.ListenAndServe(":8080", mux)
}
