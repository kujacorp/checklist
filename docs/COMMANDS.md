This gives useful commands.

# Backend Interaction

Login
```sh
curl -X POST http://localhost:3000/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your-username",
    "password": "your-password"
  }'
```

Get all todo lists
```bash
curl -X GET http://localhost:3000/api/lists \
  -H "Authorization: Bearer $JWT_TOKEN"
```

Create a new todo list
```bash
curl -X POST http://localhost:3000/api/lists \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My New Todo List",
    "description": "This is a description of the todo list."
  }'
```

Get a todo list
```sh
curl -X GET http://localhost:3000/api/lists/<list-id> \
  -H "Authorization: Bearer $JWT_TOKEN"
```

Update a todo list
```sh
curl -X PUT http://localhost:3000/api/lists/<list-id> \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Todo List Title",
    "description": "Updated description of the todo list."
  }'
```

Delete a todo list
```sh
curl -X DELETE http://localhost:3000/api/lists/<list-id> \
  -H "Authorization: Bearer $JWT_TOKEN"
```

Create a new todo in a list
```sh
curl -X POST http://localhost:3000/api/lists/<list-id>/todos \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Todo Item",
    "description": "Description of the todo item."
  }'
```

Get a todo
```sh
curl -X GET http://localhost:3000/api/todos/<todo-id> \
  -H "Authorization: Bearer $JWT_TOKEN"
```

Update a todo
```sh
curl -X PUT http://localhost:3000/api/todos/<todo-id> \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Todo Title",
    "description": "Updated description.",
    "completed": true
  }'
```

Delete a todo
```sh
curl -X DELETE http://localhost:3000/api/todos/<todo-id> \
  -H "Authorization: Bearer $JWT_TOKEN"
```
