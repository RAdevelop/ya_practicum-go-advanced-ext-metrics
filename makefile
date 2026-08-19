# Переменные
iter ?=

.PHONY: help
help: ## Показать справку
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)  make %-20s$(NC) %s\n", $$1, $$2}'

.PHONY: test
test: ## Запустить локальные тесты
	@go test ./...

.PHONY: test-v
test-v: ## Запустить тесты с подробным выводом
	@echo "$(GREEN)=== Running tests (verbose) ===$(NC)"
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
	&& echo $$ADDRESS
	@echo "$(GREEN)✅ Tests completed$(NC)"


.DEFAULT_GOAL := help