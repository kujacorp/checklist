You can host the app by building the backend and the frontend and tying them together with Nginx.

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

3. Create an Nginx configuration file (`/etc/nginx/sites-available/checklist.conf`):
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
# Set the PostgreSQL connection string
export POSTGRES_DSN="host=localhost user=postgres password=postgres dbname=postgres sslmode=disable"

# Run the server
/path/to/backend/server
```

5. In another separate terminal, enable Nginx:
```sh
sudo ln -s /etc/nginx/sites-available/checklist.conf /etc/nginx/sites-enabled/

# Test Nginx configuration
sudo nginx -t

# Restart Nginx
sudo systemctl restart nginx
```
