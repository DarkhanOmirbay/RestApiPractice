package main

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"rest/internal/data"
	"time"
)

func (app *application) Register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}
	if input.Email == "" {
		app.badRequestResponse(w, r, errors.New("email is required"))
	}
	if input.Password == "" {
		app.badRequestResponse(w, r, errors.New("password is required"))
	}
	pass_hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	_, err = app.db.Exec("INSERT INTO users(email,pass_hash) VALUES($1,$2)", input.Email, pass_hash)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	fmt.Fprintf(w, "%+v\n", input)

}
func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}
	if input.Email == "" {
		app.badRequestResponse(w, r, errors.New("email is required"))
	}
	if input.Password == "" {
		app.badRequestResponse(w, r, errors.New("password is required"))
	}
	var user data.User
	err = app.db.QueryRow("SELECT id,email,pass_hash FROM users WHERE email=$1", input.Email).Scan(&user.ID, &user.Email, &user.Password.Hash)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.serverErrorResponse(w, r, errors.New("user not found"))
		default:
			app.serverErrorResponse(w, r, err)
		}
	}
	err = bcrypt.CompareHashAndPassword(user.Password.Hash, []byte(input.Password))
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	token, err := NewToken(user, time.Hour)
	tokenHash := sha256.Sum256([]byte(token))
	_, err = app.db.Exec("INSERT INTO tokens(hash,user_id,expiry) VALUES($1,$2,$3)", tokenHash[:], user.ID, time.Now().Add(time.Hour))
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	fmt.Fprintf(w, "%+v\n", token)
}
func NewToken(user data.User, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
