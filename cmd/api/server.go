package main

import (
	"errors"
	"fmt"
	"go.temporal.io/sdk/client"
	"net/http"
	"time"
)

type expenseState string

const (
	created   expenseState = "CREATED"
	approved  expenseState = "APPROVED"
	rejected  expenseState = "REJECTED"
	completed expenseState = "COMPLETED"
)

// Server memory store (fake DB) for expenses
var (
	allExpense     = make(map[string]expenseState)
	tokenMap       = make(map[string][]byte)
	workflowClient client.Client
)

func (app *application) serve() error {
	var err error

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	app.logger.Printf("starting server on address%s", srv.Addr)

	workflowClient, err = client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		return err
	}

	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
