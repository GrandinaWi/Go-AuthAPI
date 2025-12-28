package main

import (
	"gostart/internal/app"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load(".env")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}
	a, err := app.New(dsn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("API started on :8080")
	log.Fatal(a.Run(":8080"))
}
