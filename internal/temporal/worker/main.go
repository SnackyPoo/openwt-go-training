package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"gotrain.mzanko.net/internal/temporal/workflow"
)

func main() {
	log.Println("INFO  Starting worker...")

	// Register the workflow and activities with the worker
	// The client and worker are heavyweight objects that should be created once per process
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// Worker
	w := worker.New(c, "expense", worker.Options{})

	// Workflow registration
	w.RegisterWorkflow(workflow.ExpenseWorkflow)

	// Activities registration
	w.RegisterActivity(workflow.CreateExpenseActivity)
	w.RegisterActivity(workflow.WaitForDecisionActivity)
	w.RegisterActivity(workflow.PaymentActivity)

	// Run the Worker
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
