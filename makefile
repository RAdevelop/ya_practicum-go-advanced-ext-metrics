# Переменные
## Это конечно не хранить в Git.
export POSTGRES_USER := go-advanced-ext
export POSTGRES_PASSWORD := go-advanced-ext
export POSTGRES_DB := go-advanced-ext

# Определение команды docker-compose
DOCKER_COMPOSE := docker-compose

#номер инкремента
iter ?=

# Цвета для вывода
GREEN = \033[0;32m
YELLOW = \033[0;33m
NC = \033[0m # No Color

.PHONY: help
help: ## Показать справку
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)  make %-20s$(NC) %s\n", $$1, $$2}'

.PHONY: test
test: ## Запустить локальные тесты
	@echo "$(GREEN)=== Running tests ===$(NC)"
	@go generate ./...
	@go test ./...
	@echo "$(GREEN)✅ Tests completed$(NC)"

.PHONY: test-c
test-c: ## Запустить тесты без кэширования
	@echo "$(GREEN)=== Running tests (verbose) ===$(NC)"
	@go generate ./...
	@go test ./... -count=1
	@echo "$(GREEN)✅ Tests completed$(NC)"

.PHONY: test-v
test-v: ## Запустить тесты с подробным выводом
	@echo "$(GREEN)=== Running tests (verbose) ===$(NC)"
	@go generate ./...
	@go test ./... -v
	@echo "$(GREEN)✅ Tests completed$(NC)"

.PHONY: test-vс
test-vc: ## Запустить тесты с подробным выводом без кэширования
	@echo "$(GREEN)=== Running tests (verbose) ===$(NC)"
	@go generate ./...
	@go test ./... -v -count=1
	@echo "$(GREEN)✅ Tests completed$(NC)"

.PHONY: test-iter
test-iter: ## Запустить тесты практикума (make test-iter iter=номер_задания)
	@echo "$(GREEN)=== Running tests (practicum) ===$(NC)"
	@go build -o ./cmd/server/server ./cmd/server/*.go \
	&& go build -o ./cmd/agent/agent ./cmd/agent/*.go \
	&& SERVER_PORT=8080 \
	&& ADDRESS="localhost:$${SERVER_PORT}" \
	&& TEMP_FILE="iter9.json" \
	&& ./metricstest_v2-darwin-amd64 -test.v -test.run=^TestIteration${iter}$$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-server-port=$${SERVER_PORT} \
		-source-path=. \
		-test.failfast \
		-file-storage-path=$${TEMP_FILE} \
	&& echo $$ADDRESS && rm -f $$TEMP_FILE
	@echo "$(GREEN)✅ Tests completed$(NC)"

.PHONY: test-iter10x
test-iter10x: ## Запустить тесты практикума с 10 по 14 задание они идут с БД (make test-iter iter=номер_задания)
	@echo "$(GREEN)=== Running tests (practicum) ===$(NC)"
	@go build -o ./cmd/server/server ./cmd/server/*.go \
	&& go build -o ./cmd/agent/agent ./cmd/agent/*.go \
	&& SERVER_PORT=8080 \
	&& ADDRESS="localhost:$${SERVER_PORT}" \
	&& TEMP_FILE="iter9.json" \
	&& ./metricstest_v2-darwin-amd64 -test.v -test.run=^TestIteration${iter}$$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-database-dsn='postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable' \
		-server-port=$${SERVER_PORT} \
		-source-path=. \
		-test.failfast \
		-file-storage-path=$${TEMP_FILE} \
	&& echo $$ADDRESS && rm -f $$TEMP_FILE
	@echo "$(GREEN)✅ Tests completed$(NC)"

.PHONY: unset-env
unset-env: ## Удалить ENV переменные
	@echo "$(GREEN)=== Cleaning environment ===$(NC)"
	@for var in POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB; do \
		unset $$var; \
	done
	@echo "$(GREEN)✅ Environment cleaned$(NC)"

.PHONY: down
down: ## Остановить кластер
	@echo "$(YELLOW)=== Stopping cluster ===$(NC)"
	@$(DOCKER_COMPOSE) down
	@echo "$(GREEN)✅ Cluster stopped$(NC)"

.PHONY: clean
clean: ## Остановить кластер и удалить данные (volumes)
	@echo "$(YELLOW)=== Cleaning cluster data ===$(NC)"
	@$(DOCKER_COMPOSE) down -v
	@echo "$(GREEN)✅ Cluster data cleaned$(NC)"

.PHONY: up
up: ## Запустить кластер
	@echo "$(GREEN)=== Starting cluster ===$(NC)"
	@$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)✅ Cluster started$(NC)"
	@make status

.PHONY: status
status: ## Проверить статус контейнеров
	@echo "$(GREEN)=== Cluster status ===$(NC)"
	@$(DOCKER_COMPOSE) ps

.PHONY: build
build:  ## Собрать кластер
	@make down
	@make clean
	@make up
	@make status
	@make unset-env

.PHONY: rebuild
rebuild:  ## пересобрать кластер
	@make down
	@make clean
	@make up --no-deps --build
	@make status
	@make unset-env

.DEFAULT_GOAL := help