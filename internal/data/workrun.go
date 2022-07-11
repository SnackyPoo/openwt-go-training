package data

// WorkRun is a workflow run
// In this case a Workflow will always have a single run
// Otherwise, a Workflow cron job can have multiple runs (per day, hour etc.) according to cron
type WorkRun struct {
	WorkflowId string `json:"workflowId"`
	RunId      string `json:"runId"`
}
