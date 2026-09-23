package models

import "time"

type User struct {
	ID            string    `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Email         string    `json:"email" db:"email"`
	Avatar_url    string    `json:"avatar_url" db:"avatar_url"`
	Password_hash string    `json:"-" db:"password_hash"`
	Created_at    time.Time `json:"created_at" db:"created_at"`
	Updated_at    time.Time `json:"updated_at" db:"updated_at"`
}

type UserRegisterInput struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Avatar_url string `json:"avatar_url"`
	Password   string `json:"password"`
}

type UserLoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserCreateInput struct {
	Name       string
	Email      string
	Avatar_url string
	Password   string
}
