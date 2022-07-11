# Go Training with Temporal
This sample workflow, processes an expense request. The key part of this exercise is to show how to complete an activity 
asynchronously. This is a modified Temporal Go sample.

To find out more about Temporal, go here: https://temporal.io/

## Description
* Create a new expense report. Expense report has just the ID and Status, without amount or date to make things simpler.
* Wait for the expense report to be approved. This could take an arbitrary amount of time. So the activity's `Execute` 
method has to return before it is actually approved. This is done by returning a special error so the framework knows 
the activity is not completed yet. 
* When the expense is approved (or rejected), somewhere in the world needs to be notified, and it will need to call
`client.CompleteActivity()` to tell Temporal service that that activity is now completed. 
In this sample case, the dummy server does this job. In real world, you will need to register some listener 
to the expense system or you will need to have your own polling agent to check for the expense status periodically. 
* After the wait activity is completed, it does the payment for the expense (dummy step in this sample case).

This sample relies on a dummy expense server w/ REST API to work.

## Setup the services
You need a Temporal service running. Clone the repository first.
```
git clone git@github.com:temporalio/docker-compose.git
```
Change dir to project and run the docker compose to start Temporal service.
```
cd docker-compose
```
```
docker-compose up
```
Get this repo.
```
git clone git@github.com:SnackyPoo/openwt-go-training.git
```
Start the dummy server with REST API. Note, since the expenses are just kept in memory, if you shutdown or restart the dummy server, they will be gone and workflow never completed (unless manually terminated). 
```
go run ./cmd/api
```
Start the activity worker.
```
go run ./internal/temporal/worker
```
## Use the REST API to execute the workflow
Import collection into Postman from:
```
/openwt-gw-training/postman/OpenWT Go Training.postman_collection.json
```

In Postman, start expense workflow executions by sending several requests, which will start workflow executions for them.
```
Add Expense
```
In Postman, list all added expenses to see their IDs and statuses.
```
Get All Expenses
```
Go to Temporal service web app to check all the Workflows. Click on a Workflow execution for details.
```
http://localhost:8080
```
In Postman, reject one and accept another expense, using their IDs. After you approve or reject the expense, workflow will complete.
```
Approve/Reject Expense by Id
```
In Postman, list all expenses to see their statuses, or get just one.
```
Get Expense by Id
```
You can also check the workflow statuses on Temporal service web app again.

