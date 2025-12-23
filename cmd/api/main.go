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
var jwtSecret = []byte("super-secret-key")

type contextKey string

const userIDKey contextKey = "userID"

//func getUserId(w http.ResponseWriter, r *http.Request) {
//	if r.Method != http.MethodPost {
//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//		return
//	}
//	var creds struct {
//		Username string `json:"username"`
//		Password string `json:"password"`
//	}
//	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
//		http.Error(w, "Json invalid", http.StatusBadRequest)
//		return
//	}
//	var userID int64
//	err := db.QueryRow(
//		"SELECT id FROM public.users WHERE username=$1 AND password=$2",
//		creds.Username,
//		creds.Password,
//	).Scan(&userID)
//
//	if err == sql.ErrNoRows {
//		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
//		return
//	}
//	if err != nil {
//		log.Println("login query error:", err)
//		http.Error(w, "DB error", http.StatusInternalServerError)
//		return
//	}
//	claims := jwt.MapClaims{
//		"user_id": userID,
//		"exp":     time.Now().Add(time.Hour).Unix(),
//	}
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//	tokenString, err := token.SignedString(jwtSecret)
//	if err != nil {
//		http.Error(w, "token invalid", http.StatusBadRequest)
//		return
//	}
//	json.NewEncoder(w).Encode(map[string]string{
//		"token": tokenString,
//	})
//}

func getUser(w http.ResponseWriter, r *http.Request) {

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
	var user User
	err := db.QueryRow("SELECT username,age FROM public.users WHERE id = $1", userID).Scan(&user.Username, &user.Age)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{
		"user_id":  userID,
		"username": user.Username,
		"age":      user.Age,
	})
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
func loginHandler(w http.ResponseWriter, r *http.Request, repo user.Repository) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&creds)
	u, err := repo.GetByCredentials(r.Context(), creds.Username, creds.Password)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
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
func registerHandler(w http.ResponseWriter, r *http.Request, repo user.Repository) {
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
	u, err := repo.Create(r.Context(), creds.Username, creds.Password, creds.Age)
	if err != nil {
		if errors.Is(err, user.ErrUserAlreadyExists) {
			http.Error(w, "User already exists", http.StatusConflict) // 409
			return
		}
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{
		"id": u.ID,
	})

}
func main() {
	var err error
	dsn := "postgres://postgres:Exdark123@localhost:5432/postgres?sslmode=disable"
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	userRepo := user.NewPostgres(db)
	log.Println("Connected to postgres")

	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		loginHandler(w, r, userRepo)
	})
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		registerHandler(w, r, userRepo)
	})
	mux.Handle("/user", authMiddleware(http.HandlerFunc(getUser)))
	http.ListenAndServe(":8080", mux)
}
