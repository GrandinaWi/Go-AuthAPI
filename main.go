package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	Username string `json:"username"`
	Age      int64  `json:"age"`
	Password string `json:"password"`
}

var users = map[int64]User{
	1: {
		Username: "Bellaria02",
		Age:      25,
		Password: "Exdark123",
	},
}

// в ПРОДЕ обязательно
// env-переменная
//
// длинный ключ
//
// не в коде
var jwtSecret = []byte("super-secret-key")

type contextKey string

const userIDKey contextKey = "userID"

func getUserId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Json invalid", http.StatusBadRequest)
		return
	}
	for id, user := range users {
		if user.Username == creds.Username && user.Password == creds.Password {
			claims := jwt.MapClaims{
				"user_id": id,
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
			return
		}
	}
	http.Error(w, "Invalid  not found", http.StatusNotFound)
}

func getUser(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	userID := r.Context().Value(userIDKey).(int64)
	user, ok := users[userID]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", getUserId)
	mux.Handle("/user", authMiddleware(http.HandlerFunc(getUser)))
	http.ListenAndServe(":8080", mux)
}
