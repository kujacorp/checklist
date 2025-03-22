# Development
Run `docker compose up` to host the app locally at [localhost:3000](http://localhost:3000).

We also provide a Makefile for building outside of Docker and with additional commands for testing and linting.
See more details for this in [docs/MAKEFILE.md](./docs/MAKEFILE.md).

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

3. Connect to the database (run this in the PostgreSQL container from Step 2):
```bash
psql -U postgres
```

4. Insert the user (use the hash from step 1):
```sql
INSERT INTO users (username, password_hash) VALUES ('admin', PASSWORD_HASH_FROM_STEP_ONE);
```

Access the admin page at [localhost:8080/admin](http://localhost:8080/admin).

# Hosting
See the hosting instructions at [docs/HOSTING.md](./docs/HOSTING.md) for instructions on how you may host the app.
