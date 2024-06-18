package main

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"rest/internal/data"
	"time"
)

func (app *application) CreateDoctor(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name       string `json:"Name"`
		Surname    string `json:"Surname"`
		Position   string `json:"Position"`
		Age        uint8  `json:"Age"`
		Experience int8   `json:"Experience"`
		Token      string `json:"Token"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}
	if input.Token == "" {
		app.badRequestResponse(w, r, errors.New("token is required"))

	}
	_, isAuth, err := app.CheckToken(w, r, input.Token)
	if err != nil || !isAuth {
		app.badRequestResponse(w, r, err)
	}

	if input.Name == "" {
		app.badRequestResponse(w, r, errors.New("Name is required"))
	}
	if input.Surname == "" {
		app.badRequestResponse(w, r, errors.New("Surname is required"))
	}
	if input.Position == "" {
		app.badRequestResponse(w, r, errors.New("position is required"))
	}
	if input.Age <= 0 {
		app.badRequestResponse(w, r, errors.New("age is required"))
	}
	if input.Experience <= 0 {
		app.badRequestResponse(w, r, errors.New("exp is required"))
	}
	_, err = app.db.Exec("INSERT INTO doctors(name,surname,position,age,experience) VALUES($1,$2,$3,$4,$5)", input.Name, input.Surname, input.Position, input.Age, input.Experience)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	fmt.Fprintf(w, "%+v\n", input)
}
func (app *application) CheckToken(w http.ResponseWriter, r *http.Request, token string) (int64, bool, error) {
	tokenHash := sha256.Sum256([]byte(token))
	row := app.db.QueryRow("SELECT id,email,pass_hash FROM users INNER JOIN tokens t ON users.id=t.user_id WHERE t.hash=$1 AND t.expiry>$2", tokenHash[:], time.Now())
	var user data.User
	err := row.Scan(&user.ID, &user.Email, &user.Password.Hash)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.serverErrorResponse(w, r, errors.New("user not found"))
			return 0, false, err
		default:
			app.serverErrorResponse(w, r, err)
			return 0, false, err
		}
	}
	log.Println("token checked")
	return user.ID, true, nil
}
func (app *application) GetAllDoctors(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Token      string `json:"Token"`
		Experience int8   `json:"Experience"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	if input.Token == "" {
		app.badRequestResponse(w, r, errors.New("token is required"))

	}
	_, isAuth, err := app.CheckToken(w, r, input.Token)
	if err != nil || !isAuth {
		app.badRequestResponse(w, r, err)
	}
	rows, err := app.db.Query("SELECT * FROM doctors WHERE experience=$1 ORDER BY id", input.Experience)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	defer rows.Close()
	doctors := []data.Doctor{}
	for rows.Next() {
		var doctor data.Doctor
		err := rows.Scan(&doctor.ID, &doctor.Name, &doctor.Surname, &doctor.Position, &doctor.Age, &doctor.Experience)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}
		doctors = append(doctors, doctor)
	}
	if err = rows.Err(); err != nil {
		app.serverErrorResponse(w, r, err)
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"doctors": doctors}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)

	}
}
func (app *application) GetDoctorByID(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	var doctor data.Doctor
	row := app.db.QueryRow("SELECT * FROM doctors WHERE id=$1", id)
	err = row.Scan(&doctor.ID, &doctor.Name, &doctor.Surname, &doctor.Position, &doctor.Age, &doctor.Experience)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"doctor": doctor}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)

	}
}
func (app *application) UpdateDoctor(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	var doctor data.Doctor
	row := app.db.QueryRow("SELECT * FROM doctors WHERE id=$1", id)
	err = row.Scan(&doctor.ID, &doctor.Name, &doctor.Surname, &doctor.Position, &doctor.Age, &doctor.Experience)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	var input struct {
		Name       *string `json:"Name"`
		Surname    *string `json:"Surname"`
		Position   *string `json:"Position"`
		Age        *uint8  `json:"Age"`
		Experience *int8   `json:"Experience"`
	}
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}
	if input.Name != nil {
		doctor.Name = *input.Name
	}
	if input.Surname != nil {
		doctor.Surname = *input.Surname
	}
	if input.Position != nil {
		doctor.Position = *input.Position
	}
	if input.Age != nil {
		doctor.Age = *input.Age
	}
	if input.Experience != nil {
		doctor.Experience = *input.Experience
	}
	_, err = app.db.Exec("UPDATE doctors SET name=$1,surname=$2,position=$3,age=$4,experience=$5 WHERE id=$6", doctor.Name, doctor.Surname, doctor.Position, doctor.Age, doctor.Experience, id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"doctor": doctor}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)

	}

}
func (app *application) DeleteDoctor(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	_, err = app.db.Exec("DELETE FROM doctors WHERE id=$1", id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	fmt.Fprint(w, "doctor deleted")
}
