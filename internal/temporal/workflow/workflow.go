package workflow

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

var expenseServerUrl = "http://localhost:4000"

// ExpenseWorkflow Workflow definition
func ExpenseWorkflow(ctx workflow.Context, expenseID string) (result string, err error) {

	logger := workflow.GetLogger(ctx)

	// Step 1: Create a new expense
	// Set a timeout at 10 minutes, otherwise fail the activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)

	logger.Info("Create Expense")

	err = workflow.ExecuteActivity(ctx1, CreateExpenseActivity, expenseID).Get(ctx1, nil)
	if err != nil {
		logger.Error("Failed to create Expense Report", "Error", err)
		return "", err
	}

	logger.Info("Expense created")

	// Step 2: Wait for the expense to be approved (or rejected)
	// We wait for human to approve the request for 10 minutes
	// If a human fails to approve within 10 minutes, Temporal will mark the activity as failure
	ao = workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx2 := workflow.WithActivityOptions(ctx, ao)
	var status string
	err = workflow.ExecuteActivity(ctx2, WaitForDecisionActivity, expenseID).Get(ctx2, &status)
	if err != nil {
		return "", err
	}

	if status != "APPROVED" {
		logger.Info("Workflow completed.", "ExpenseStatus", status)
		return "COMPLETED (" + status + ")", nil
	}

	// Step 3: Request payment for the expense
	err = workflow.ExecuteActivity(ctx2, PaymentActivity, expenseID).Get(ctx2, nil)
	if err != nil {
		logger.Info("Workflow completed with failed payment.", "Error", err)
		return "", err
	}

	logger.Info("Workflow completed in full, with expense payment successfully completed.")
	return "COMPLETED (APPROVED)", nil
}
