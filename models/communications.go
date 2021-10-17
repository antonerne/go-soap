package models

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type ErrorMessage struct {
	ErrorType  string `json:"errortype"`
	StatusCode int32  `json:"status"`
	Message    string `json:"message"`
}

func (em *ErrorMessage) String() string {
	return fmt.Sprintf("(%s) %s", em.ErrorType, em.Message)
}

type JwtClaims struct {
	Id         string    `json:"id"`
	Email      string    `json:"email"`
	Roles      []string  `json:"roles"`
	Exp        int64     `json:"exp"`
	Expires    time.Time `json:"expires"`
	MustChange bool      `json:"mustchange"`
	Locked     bool      `json:"locked"`
	Uuid       string    `json:"uuid"`
	jwt.StandardClaims
}

type LoginResponse struct {
	Token string `json:"token"`
}
