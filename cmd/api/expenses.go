package main

import (
	"context"
	"fmt"
	"gotrain.mzanko.net/internal/data"
	"gotrain.mzanko.net/internal/temporal/workflow"
	"sort"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
	"net/http"
)

func (app *application) starterHandler(w http.ResponseWriter, r *http.Request) {
	expenseID := uuid.New()

	c, err := client.Dial(client.Options{HostPort: client.DefaultHostPort})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "expense_" + expenseID,
		TaskQueue: "expense",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, workflow.ExpenseWorkflow, expenseID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	workRun := &data.WorkRun{
		WorkflowId: we.GetID(),
		RunId:      we.GetRunID(),
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"workflowRun": workRun}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) createExpenseHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Id string `json:"id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	_, ok := allExpense[input.Id]
	if ok {
		app.duplicateExpenseResponse(w, r)
		return
	}

	allExpense[input.Id] = created

	_, err = fmt.Fprintf(w, "SUCCEED")
	if err != nil {
		return
	}

	app.logger.Printf("Created new expense id: %s\n", input.Id)
}

func (app *application) listExpensesHandler(w http.ResponseWriter, r *http.Request) {
	var keys []string
	for k := range allExpense {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var expenses []*data.Expense

	for _, id := range keys {
		state := allExpense[id]
		expenses = append(expenses, &data.Expense{
			Id:    id,
			State: string(state),
		})
	}

	err := app.writeJSON(w, http.StatusOK, envelope{"expenses": expenses}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) showExpenseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	state, ok := allExpense[id]
	if !ok {
		app.notFoundResponse(w, r)
		return
	}

	expense := &data.Expense{
		Id:    id,
		State: string(state),
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"expense": expense}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) callbackHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	currState, ok := allExpense[id]
	if !ok {
		_, _ = fmt.Fprint(w, "ERROR:INVALID_ID")
		return
	}
	if currState != created {
		_, _ = fmt.Fprint(w, "ERROR:INVALID_STATE")
		return
	}

	if err != nil {
		// Handle error here via logging and then return
		_, _ = fmt.Fprint(w, "ERROR:INVALID_FORM_DATA")
		return
	}

	taskToken := r.PostFormValue("task_token")
	fmt.Printf("Registered callback for ID=%s, token=%s\n", id, taskToken)
	tokenMap[id] = []byte(taskToken)
	_, _ = fmt.Fprint(w, "SUCCEED")
}

func (app *application) actionExpenseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	actionType, err := app.readActionParameter(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	oldState, ok := allExpense[id]
	if !ok {
		_, _ = fmt.Fprint(w, "ERROR:INVALID_ID")
		return
	}

	switch actionType {
	case "approve":
		allExpense[id] = approved
	case "reject":
		allExpense[id] = rejected
	case "payment":
		allExpense[id] = completed
	}

	_, _ = fmt.Fprint(w, "SUCCEED")

	if oldState == created && (allExpense[id] == approved || allExpense[id] == rejected) {
		// report state change
		notifyExpenseStateChange(id, string(allExpense[id]))
	}

	fmt.Printf("Set state for %s from %s to %s.\n", id, oldState, allExpense[id])
}

func notifyExpenseStateChange(id, state string) {
	token, ok := tokenMap[id]
	if !ok {
		fmt.Printf("Invalid id:%s\n", id)
		return
	}
	err := workflowClient.CompleteActivity(context.Background(), token, state, nil)
	if err != nil {
		fmt.Printf("Failed to complete activity with error: %+v\n", err)
	} else {
		fmt.Printf("Successfully complete activity: %s\n", token)
	}
}
