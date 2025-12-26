package user

import "errors"

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Age      int    `json:"age"`
}

var ErrUserAlreadyExists = errors.New("user already exists")
