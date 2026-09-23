package handlers

import (
	"os"
)

var JWTSecret = []byte(os.Getenv("JWT_SECRET"))
