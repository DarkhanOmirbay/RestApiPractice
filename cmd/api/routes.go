package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() *httprouter.Router {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodPost, "/api/v1/register", app.Register)
	router.HandlerFunc(http.MethodPost, "/api/v1/login", app.Login)

	router.HandlerFunc(http.MethodPost, "/api/v1/doctors", app.CreateDoctor)
	router.HandlerFunc(http.MethodPost, "/api/v1/getall", app.GetAllDoctors)

	router.HandlerFunc(http.MethodGet, "/api/v1/doctors/:id", app.GetDoctorByID)
	router.HandlerFunc(http.MethodPut, "/api/v1/doctors/:id", app.UpdateDoctor)
	router.HandlerFunc(http.MethodDelete, "/api/v1/doctors/:id", app.DeleteDoctor)

	return router
}
