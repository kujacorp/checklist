# Development
Run `docker compose up` to host the app locally at [localhost:3000](http://localhost:3000).

## Admin Setup

After the web application is initialized, create a user to access the admin tools.

1. Generate a bcrypt password hash, we provide a tool:
```bash
# Replace 'your_secure_password' with an actual secure password
cd backend
go run tools/generate_password.go your_secure_password
```

2. Connect to the PostgreSQL container:
```bash
docker compose exec postgres bash
```

3. Connect to the database:
```bash
psql -U postgres
```

4. Insert the user (use the hash from step 1):
```sql
INSERT INTO users (username, password_hash)
VALUES ('admin', '$2y$05$...');
```
