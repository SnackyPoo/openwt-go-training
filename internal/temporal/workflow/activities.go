package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"go.temporal.io/sdk/activity"
)

func CreateExpenseActivity(ctx context.Context, expenseID string) error {
	if len(expenseID) == 0 {
		return errors.New("expense id is empty")
	}
	createExpensePayload := struct {
		Id string `json:"id"`
	}{
		Id: expenseID,
	}

	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(createExpensePayload)

	req, err := http.NewRequest(http.MethodPost, expenseServerUrl+"/v1/expenses", payload)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if string(body) == "SUCCEED" {
		activity.GetLogger(ctx).Info("Expense created.", "ExpenseID", expenseID)
		return nil
	}

	return errors.New(string(body))
}

// WaitForDecisionActivity waits for the expense decision. This activity will complete asynchronously. When this method
// returns error activity.ErrResultPending, the Temporal Go SDK recognize this error, and won't mark this activity
// as failed or completed. The Temporal server will wait until Client.CompleteActivity() is called or timeout happened
// whichever happen first. In this sample case, the CompleteActivity() method is called by our dummy expense server when
// the expense is approved.
func WaitForDecisionActivity(ctx context.Context, expenseID string) (string, error) {
	if len(expenseID) == 0 {
		return "", errors.New("expense id is empty")
	}

	logger := activity.GetLogger(ctx)

	// save current activity info so, it can be completed asynchronously when expense is approved/rejected
	activityInfo := activity.GetInfo(ctx)
	formData := url.Values{}
	formData.Add("task_token", string(activityInfo.TaskToken))

	registerCallbackURL := expenseServerUrl + "/v1/workflow/registerCallback/" + expenseID
	resp, err := http.PostForm(registerCallbackURL, formData)

	if err != nil {
		logger.Info("waitForDecisionActivity failed to register callback.", "Error", err)
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return "", err
	}

	status := string(body)
	if status == "SUCCEED" {
		// register callback succeed
		logger.Info("Successfully registered callback.", "ExpenseID", expenseID)

		// ErrActivityResultPending is returned from activity's execution to indicate the activity is not completed when it returns.
		// activity will be completed asynchronously when Client.CompleteActivity() is called.
		return "", activity.ErrResultPending
	}

	logger.Warn("Register callback failed.", "ExpenseStatus", status)
	return "", fmt.Errorf("register callback failed status:%s", status)
}

func PaymentActivity(ctx context.Context, expenseID string) error {
	if len(expenseID) == 0 {
		return errors.New("expense id is empty")
	}

	resp, err := http.Get(expenseServerUrl + "/v1/expenses/" + expenseID + "/payment")
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}

	if string(body) == "SUCCEED" {
		activity.GetLogger(ctx).Info("paymentActivity succeed", "ExpenseID", expenseID)
		return nil
	}

	return errors.New(string(body))
}
