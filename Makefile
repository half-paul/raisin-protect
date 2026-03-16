# Raisin Protect — Makefile
# Common development tasks

# Docker container name for postgres (matches docker-compose)
DB_CONTAINER ?= rp-postgres
DB_USER ?= rp
DB_NAME ?= raisin_protect

# Helper: run SQL file inside the postgres container
# Files are mounted at /docker-entrypoint-initdb.d/ by docker-compose
DOCKER_PSQL = docker exec -i $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -q

.PHONY: help seed seed-frameworks migrate test build

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# --------------------------------------------------------------------------
# Database
# --------------------------------------------------------------------------

migrate: ## Run all migrations in order (via Docker)
	@echo "Running migrations..."
	@for f in $$(ls db/migrations/*.sql | sort); do \
		echo "  → $$f"; \
		$(DOCKER_PSQL) < "$$f" 2>&1 | grep -v "^$$" || true; \
	done
	@echo "Migrations complete."

seed: ## Run full seed (demo org, users, frameworks, controls, mappings)
	@echo "Running seed.sql..."
	$(DOCKER_PSQL) < db/seeds/seed.sql
	@echo "Seed complete."

seed-frameworks: ## Seed frameworks only (PCI DSS v3.2.1 + HIPAA, NIST CSF, CIS Controls, SOX)
	@echo "Seeding PCI DSS v3.2.1..."
	$(DOCKER_PSQL) < db/migrations/062_pci_dss_v321.sql
	@echo "Seeding additional frameworks (HIPAA, NIST CSF, CIS Controls, SOX)..."
	$(DOCKER_PSQL) < db/migrations/063_additional_frameworks.sql
	@echo "Framework seeding complete."

seed-all: seed seed-frameworks ## Run all seeds (full seed + additional frameworks)

db-reset: ## Reset database (drop and recreate, run migrations + seeds)
	@echo "Dropping and recreating database..."
	docker exec -i $(DB_CONTAINER) psql -U $(DB_USER) -d postgres -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	docker exec -i $(DB_CONTAINER) psql -U $(DB_USER) -d postgres -c "CREATE DATABASE $(DB_NAME);"
	@$(MAKE) migrate
	@$(MAKE) seed-all

# --------------------------------------------------------------------------
# Build & Test
# --------------------------------------------------------------------------

build: ## Build the Go API
	cd api && go build ./...

test: ## Run all Go tests
	cd api && go test ./...

test-v: ## Run all Go tests with verbose output
	cd api && go test -v ./...

test-cover: ## Run tests with coverage
	cd api && go test -v -cover ./...

vet: ## Run Go vet
	cd api && go vet ./...

# --------------------------------------------------------------------------
# Docker
# --------------------------------------------------------------------------

up: ## Start all services
	docker-compose up -d

down: ## Stop all services
	docker-compose down

logs: ## Tail all service logs
	docker-compose logs -f

ps: ## Show service status
	docker-compose ps
