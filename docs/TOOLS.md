# Tools

## Generating Hashed Passwords

Generate a bcrypt password hash with:
```bash
# Replace 'your_secure_password' with an actual secure password
cd backend
go run tools/generate_password.go your_secure_password
```


## Interacting with the Backend

We provide several command-line tools to interact with the application:

### Interactive Client

An interactive shell client that allows you to authenticate and check view counts:

```bash
cd backend
go run tools/client.go
```

Once running, you'll be prompted to log in with your username and password. After authentication, you can use these commands:
- `views` - Shows the current view count from the backend
- `echo [text]` - Displays the provided text
- `type [command]` - Shows information about a command
- `exit` - Exits the client

### Shell Scripts

We also have standalone shell scripts:

#### Authentication Script

```bash
# Run this with 'source' to set the JWT in your current shell
source bin/checklist-login
```

This script will prompt for your username and password, then authenticate with the server and set the `CHECKLIST_JWT` environment variable.

#### View Count Script

```bash
bin/checklist-get-views
```

This script fetches the current view count from the backend. It requires the `CHECKLIST_JWT` environment variable to be set (by running the login script first).

Example workflow:
```bash
# Login first
source bin/checklist-login

# Now you can check views as many times as needed
bin/checklist-get-views
```

For automation, you can also use:
```bash
# One-liner to login and capture the JWT
eval $(bin/checklist-login)

# Now use the JWT in scripts
bin/checklist-get-views
```
