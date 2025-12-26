package user

import (
	"context"
	"database/sql"
	"strings"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{db: db}
}

func (r *Postgres) GetByID(ctx context.Context, id int64) (*User, error) {
	var u User

	err := r.db.QueryRowContext(ctx, "SELECT id, username, age, password FROM users WHERE id = $1", id).Scan(&u.ID, &u.Username, &u.Age, &u.Password)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Postgres) GetByCredentials(ctx context.Context, username string, password string) (*User, error) {
	var u User
	err := r.db.QueryRowContext(ctx, "SELECT id, username, age, password FROM users WHERE username = $1 AND password = $2", username, password).Scan(&u.ID, &u.Username, &u.Age, &u.Password)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
func (r *Postgres) Create(ctx context.Context, username string, password string, age int64) (*User, error) {
	var u User
	err := r.db.QueryRowContext(ctx, "INSERT INTO users (username, age, password) VALUES ($1, $2, $3) RETURNING id", username, age, password).Scan(&u.ID)
	if err != nil {
		return nil, err
	}
	if err != nil {
		// ловим UNIQUE violation
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}
	return &u, nil
}
