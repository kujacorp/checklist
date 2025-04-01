You can host the app by building the backend and the frontend and tying them together with Nginx.

# Pre-Requisites
You will need:
- A server with Go and Node.js installed
- A PostgreSQL instance and its connection info (host, user, password, dbname)

# Instructions

1. Build the backend
```bash
# From the root directory
cd backend
go build -o server main.go
```

2. Build the frontend:
```bash
# From the root directory
cd frontend
npm run build
```

3. Create an Nginx configuration file (`/etc/nginx/conf.d/checklist.conf`):
```conf
server {
    listen 80;
    server_name yourdomain.com;  # Replace with your domain

    # Serve frontend static files
    location / {
        root /path/to/your/frontend/dist;
        try_files $uri $uri/ /index.html;
    }

    # Proxy API requests to Go backend
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

4. Set up environment variables and run the backend:
```bash
# Set the PostgreSQL connection string, replace values with the details of your Postgres setup
export POSTGRES_DSN="host=localhost user=postgres password=postgres dbname=postgres sslmode=disable"

# Generate a random JWT secret key
export JWT_SECRET_KEY=$(openssl rand -base64 32)

# Run the server
/path/to/backend/server
```

5. In another separate terminal, restart Nginx to load the config in step 3:
```sh
# Test Nginx configuration
sudo nginx -t

# Restart Nginx
sudo systemctl restart nginx
```
