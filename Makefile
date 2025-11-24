.PHONY: help build up down restart logs clean backup restore

# Variables
COMPOSE_FILE ?= docker-compose.prod.yml
PROJECT_NAME = whatsapp

# Colors
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput -Txterm sgr0)

help: ## Show this help message
	@echo '$(GREEN)Available commands:$(RESET)'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(RESET) %s\n", $$1, $$2}'

build: ## Build all Docker images
	@echo "$(GREEN)Building Docker images...$(RESET)"
	docker-compose -f $(COMPOSE_FILE) build

build-no-cache: ## Build all Docker images without cache
	@echo "$(GREEN)Building Docker images (no cache)...$(RESET)"
	docker-compose -f $(COMPOSE_FILE) build --no-cache

up: ## Start all services
	@echo "$(GREEN)Starting services...$(RESET)"
	docker-compose -f $(COMPOSE_FILE) up -d
	@echo "$(GREEN)Services started! Check status with: make status$(RESET)"

down: ## Stop all services
	@echo "$(YELLOW)Stopping services...$(RESET)"
	docker-compose -f $(COMPOSE_FILE) down

restart: ## Restart all services
	@echo "$(YELLOW)Restarting services...$(RESET)"
	docker-compose -f $(COMPOSE_FILE) restart

stop: ## Stop all services without removing containers
	@echo "$(YELLOW)Stopping services...$(RESET)"
	docker-compose -f $(COMPOSE_FILE) stop

start: ## Start stopped services
	@echo "$(GREEN)Starting services...$(RESET)"
	docker-compose -f $(COMPOSE_FILE) start

status: ## Show services status
	@docker-compose -f $(COMPOSE_FILE) ps

logs: ## Show logs for all services
	docker-compose -f $(COMPOSE_FILE) logs -f --tail=100

logs-backend: ## Show backend logs
	docker-compose -f $(COMPOSE_FILE) logs -f backend --tail=100

logs-frontend: ## Show frontend logs
	docker-compose -f $(COMPOSE_FILE) logs -f frontend --tail=100

logs-nginx: ## Show nginx logs
	docker-compose -f $(COMPOSE_FILE) logs -f nginx --tail=100

logs-mongo: ## Show MongoDB logs
	docker-compose -f $(COMPOSE_FILE) logs -f mongo --tail=100

shell-backend: ## Open shell in backend container
	docker-compose -f $(COMPOSE_FILE) exec backend sh

shell-frontend: ## Open shell in frontend container
	docker-compose -f $(COMPOSE_FILE) exec frontend sh

shell-mongo: ## Open MongoDB shell
	docker-compose -f $(COMPOSE_FILE) exec mongo mongosh

clean: ## Remove all containers, volumes, and images
	@echo "$(YELLOW)Cleaning up Docker resources...$(RESET)"
	docker-compose -f $(COMPOSE_FILE) down -v
	docker system prune -af

clean-volumes: ## Remove all volumes (WARNING: deletes all data)
	@echo "$(YELLOW)WARNING: This will delete all data!$(RESET)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose -f $(COMPOSE_FILE) down -v; \
		echo "$(GREEN)Volumes removed$(RESET)"; \
	fi

backup: ## Backup MongoDB and volumes
	@echo "$(GREEN)Creating backup...$(RESET)"
	@mkdir -p backups
	@docker-compose -f $(COMPOSE_FILE) exec -T mongo \
		mongodump --username=$${MONGO_USER} --password=$${MONGO_PASS} \
		--authenticationDatabase=admin --db=whatsapp_production \
		--archive > backups/mongodb_backup_$$(date +%Y%m%d_%H%M%S).archive
	@echo "$(GREEN)Backup completed!$(RESET)"

restore: ## Restore MongoDB from backup (usage: make restore FILE=backups/mongodb_backup_YYYYMMDD_HHMMSS.archive)
	@if [ -z "$(FILE)" ]; then \
		echo "$(YELLOW)Usage: make restore FILE=backups/mongodb_backup_YYYYMMDD_HHMMSS.archive$(RESET)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Restoring from $(FILE)...$(RESET)"
	docker-compose -f $(COMPOSE_FILE) stop backend
	docker-compose -f $(COMPOSE_FILE) exec -T mongo \
		mongorestore --username=$${MONGO_USER} --password=$${MONGO_PASS} \
		--authenticationDatabase=admin --drop \
		--archive < $(FILE)
	docker-compose -f $(COMPOSE_FILE) start backend
	@echo "$(GREEN)Restore completed!$(RESET)"

health: ## Check health of all services
	@echo "$(GREEN)Checking service health...$(RESET)"
	@echo "Nginx: $$(curl -s -o /dev/null -w "%{http_code}" http://localhost/health)"
	@echo "Backend: $$(docker-compose -f $(COMPOSE_FILE) exec -T backend curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/health || echo 'N/A')"
	@echo "MongoDB: $$(docker-compose -f $(COMPOSE_FILE) exec -T mongo mongosh --quiet --eval "db.adminCommand('ping').ok" 2>/dev/null || echo '0')"

deploy: build-no-cache down up ## Full deployment (build, stop, start)
	@echo "$(GREEN)Deployment completed!$(RESET)"
	@make health

update: ## Update and restart services (zero-downtime)
	@echo "$(GREEN)Updating services...$(RESET)"
	docker-compose -f $(COMPOSE_FILE) build
	docker-compose -f $(COMPOSE_FILE) up -d --no-deps --build backend
	docker-compose -f $(COMPOSE_FILE) up -d --no-deps --build frontend
	docker-compose -f $(COMPOSE_FILE) up -d --no-deps --build nginx
	@echo "$(GREEN)Update completed!$(RESET)"

stats: ## Show resource usage statistics
	docker stats --no-stream

setup-ssl: ## Setup self-signed SSL certificate (development only)
	@echo "$(GREEN)Creating self-signed SSL certificate...$(RESET)"
	@mkdir -p nginx/ssl
	@openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
		-keyout nginx/ssl/key.pem \
		-out nginx/ssl/cert.pem \
		-subj "/C=ID/ST=Jakarta/L=Jakarta/O=Development/CN=localhost"
	@echo "$(GREEN)SSL certificate created!$(RESET)"

setup-env: ## Setup production environment file
	@if [ ! -f .env.production ]; then \
		cp .env.production.example .env.production; \
		echo "$(GREEN).env.production created. Please edit it with your settings!$(RESET)"; \
	else \
		echo "$(YELLOW).env.production already exists$(RESET)"; \
	fi

install: setup-env setup-ssl ## Initial setup (env + SSL)
	@echo "$(GREEN)Initial setup completed!$(RESET)"
	@echo "$(YELLOW)Remember to edit .env.production with your settings!$(RESET)"
	@echo "$(YELLOW)For production, use proper SSL certificates (Let's Encrypt)$(RESET)"

setup: ## Interactive setup wizard
	@echo "╔══════════════════════════════════════════════════════════════╗"
	@echo "║                                                              ║"
	@echo "║          WhatsApp API - Interactive Setup Wizard            ║"
	@echo "║                                                              ║"
	@echo "╚══════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "$(GREEN)This wizard will help you configure your deployment.$(RESET)"
	@echo ""
	@./scripts/setup-wizard.sh

ssl-letsencrypt: ## Setup Let's Encrypt SSL certificate
	@if [ -z "$(DOMAIN)" ]; then \
		read -p "Enter your domain (e.g., yourdomain.com): " domain; \
	else \
		domain=$(DOMAIN); \
	fi; \
	echo "$(GREEN)Setting up Let's Encrypt for $$domain...$(RESET)"; \
	docker-compose -f $(COMPOSE_FILE) stop nginx; \
	sudo certbot certonly --standalone -d $$domain; \
	sudo cp /etc/letsencrypt/live/$$domain/fullchain.pem nginx/ssl/cert.pem; \
	sudo cp /etc/letsencrypt/live/$$domain/privkey.pem nginx/ssl/key.pem; \
	sudo chmod 644 nginx/ssl/*.pem; \
	docker-compose -f $(COMPOSE_FILE) start nginx; \
	echo "$(GREEN)SSL certificate installed!$(RESET)"

configure-domain: ## Configure domain in nginx config
	@if [ -z "$(DOMAIN)" ]; then \
		read -p "Enter your domain (e.g., yourdomain.com): " domain; \
	else \
		domain=$(DOMAIN); \
	fi; \
	echo "$(GREEN)Configuring domain: $$domain$(RESET)"; \
	sed -i.bak "s/__DOMAIN__/$$domain/g" nginx/nginx.production.conf; \
	sed -i.bak "s/CORS_ALLOWED_ORIGIN=.*/CORS_ALLOWED_ORIGIN=https:\/\/$$domain/g" .env.production; \
	echo "$(GREEN)Domain configured!$(RESET)"
