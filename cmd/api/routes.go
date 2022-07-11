package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/workflow/starter", app.starterHandler)
	router.HandlerFunc(http.MethodPost, "/v1/workflow/registerCallback/:id", app.callbackHandler)

	router.HandlerFunc(http.MethodPost, "/v1/expenses", app.createExpenseHandler)
	router.HandlerFunc(http.MethodGet, "/v1/expenses", app.listExpensesHandler)
	router.HandlerFunc(http.MethodGet, "/v1/expenses/:id", app.showExpenseHandler)
	router.HandlerFunc(http.MethodGet, "/v1/expenses/:id/:action", app.actionExpenseHandler)

	return router
}
