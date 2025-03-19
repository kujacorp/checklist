.PHONY: all build build-backend build-frontend clean lint format

all: build

build: build-backend build-frontend

build-backend:
	cd backend && go build -o server main.go

build-frontend:
	cd frontend && npm run build

clean:
	rm -f backend/server
	rm -rf frontend/dist

lint: lint-frontend

lint-frontend:
	cd frontend && npm run lint

format: format-backend format-frontend

format-backend:
	cd backend/ && go fmt ./...

format-frontend:
	cd frontend && npx prettier --write .
