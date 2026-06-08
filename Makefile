.PHONY: help up down logs migrate seed api web build typecheck

help:
	@echo "Targets:"
	@echo "  up        - start MySQL via docker-compose"
	@echo "  down      - stop MySQL"
	@echo "  logs      - tail MySQL logs"
	@echo "  migrate   - run DB migrations"
	@echo "  api       - run Go backend on :8080"
	@echo "  web       - run Vite frontend on :5173"
	@echo "  build     - build both backend and frontend"
	@echo "  typecheck - typecheck frontend"

up:
	docker compose up -d
	@echo "Waiting for MySQL to be healthy..."
	@until docker compose ps mysql --format json | grep -q '"Health":"healthy"'; do sleep 2; done
	@echo "MySQL is up. Run 'make migrate' next."

down:
	docker compose down

logs:
	docker compose logs -f mysql

migrate:
	cd backend && go run ./cmd/api migrate up

seed:
	cd backend && go run ./cmd/api migrate up

api:
	cd backend && go run ./cmd/api serve

web:
	cd frontend && npm run dev

build:
	cd backend && go build -o ../bin/api ./cmd/api
	cd frontend && npm run build

typecheck:
	cd frontend && npm run typecheck
