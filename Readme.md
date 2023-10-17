# Task Management System

## Pre-Requisite

* Postgres server with config as in code

## Starting the server
```
docker build -t task-manager .
docker run -p 8080:8080 task-manager
```

## Testing the endpoints

### POST /tasks

This endpoint allows users to create a new task. Task needs to be stored in the database. Initially mark status as incomplete. Send the task to the task channel for processing

```
curl --location 'http://localhost:8080/tasks' --header 'Content-Type: application/json' --header 'API-Key: secretkey' --data '{ "title":"Task1", "description": "This task performs operation A" }'
```
### GET /tasks

Retrieves list of all tasks

```
curl --location 'http://localhost:8080/tasks' --header 'API-Key: secretkey'
```

### GET /tasks/:id/status

Fetch status of the task

```
curl --location 'http://localhost:8080/tasks/1/status'  --header 'API-Key: secretkey'
``` 