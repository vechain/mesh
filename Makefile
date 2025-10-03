# VeChain Mesh API Makefile

.PHONY: help build test-unit test-unit-coverage test-unit-coverage-threshold test-unit-coverage-threshold-custom test-unit-coverage-html test-e2e test-e2e-verbose test-e2e-full test-e2e-vip180 test-e2e-vip180-full test-e2e-call test-e2e-call-full clean docker-build docker-up docker-down docker-logs docker-clean docker-solo-up docker-solo-down docker-solo-logs mesh-cli-build mesh-cli-check-data mesh-cli-check-construction

# Default target
help:
	@echo "VeChain Mesh API - Available commands:"
	@echo ""
	@echo "Docker commands:"
	@echo "  docker-build     - Build the Docker image"
	@echo "  docker-up        - Start services in testnet mode"
	@echo "  docker-down      - Stop services"
	@echo "  docker-logs      - View service logs"
	@echo "  docker-clean     - Remove containers and images"
	@echo "  docker-solo-up   - Start services in solo mode"
	@echo "  docker-solo-down - Stop solo mode services"
	@echo "  docker-solo-logs - View solo mode logs"
	@echo ""
	@echo "Mesh CLI validation commands:"
	@echo "  mesh-cli-build                    - Build mesh-cli Docker image"
	@echo "  mesh-cli-check-data-solo          - Validate Data API on solo network (recommended)"
	@echo "  mesh-cli-check-construction-solo  - Validate Construction API on solo network (recommended)"
	@echo "  mesh-cli-check-data ENV=<env>     - Validate Data API for specific environment (solo|test|main)"
	@echo "  mesh-cli-check-construction ENV=<env> - Validate Construction API for specific environment (solo|test|main)"
	@echo "    ⚠️  Note: test and main environments are WIP"
	@echo ""
	@echo "Development commands:"
	@echo "  build - Build the Go binary"
	@echo "  test-unit - Run unit tests (excludes e2e tests)"
	@echo "  test-unit-coverage - Run unit tests with coverage report"
	@echo "  test-unit-coverage-threshold - Run unit tests and check coverage threshold"
	@echo "  test-unit-coverage-threshold-custom - Run unit tests with custom threshold (use THRESHOLD=75)"
	@echo "  test-unit-coverage-html - Run unit tests and generate HTML coverage report"
	@echo "  test-e2e - Run e2e tests (requires solo mode server)"
	@echo "  test-e2e-verbose - Run e2e tests with verbose output"
	@echo "  test-e2e-full - Full e2e test cycle (start solo, test, stop solo)"
	@echo "  test-e2e-vip180 - Run VIP180 e2e test (requires solo mode server)"
	@echo "  test-e2e-vip180-full - Full VIP180 e2e test cycle (start solo, test, stop solo)"
	@echo "  test-e2e-call - Run Call service e2e test (requires solo mode server)"
	@echo "  test-e2e-call-full - Full Call service e2e test cycle (start solo, test, stop solo)"
	@echo "  clean - Clean Go build artifacts and cache"
	@echo ""
	@echo "Utilities:"
	@echo "  help - Show this help message"

# Development commands
build:
	go build -o mesh-server .

test-unit:
	go test $(shell go list ./... | grep -v /tests | grep -v /scripts)

test-unit-coverage:
	@echo "Generating coverage report..."
	@if ! go test -coverprofile=coverage.out $(shell go list ./... | grep -v /tests | grep -v /scripts); then \
		echo "❌ Tests failed! Cannot generate coverage report."; \
		exit 1; \
	fi
	@go tool cover -func=coverage.out | grep -v "_test.go\|mock_client.go|main.go"

test-unit-coverage-threshold:
	@$(MAKE) test-unit-coverage-threshold-custom THRESHOLD=83.6

test-unit-coverage-threshold-custom:
	@echo "Generating coverage report with custom threshold check..."
	@if [ -z "$(THRESHOLD)" ]; then \
		echo "❌ Please specify THRESHOLD (e.g., make test-unit-coverage-threshold-custom THRESHOLD=75)"; \
		exit 1; \
	fi; \
	if ! go test -coverprofile=coverage.out $(shell go list ./... | grep -v /tests | grep -v /scripts | grep -v /common/vip180/contracts); then \
		echo "❌ Tests failed! Cannot generate coverage report."; \
		exit 1; \
	fi; \
	grep -v "/_test\.go\|/mock_client\.go\|/main\.go" coverage.out > coverage_filtered.out; \
	coverage=$$(go tool cover -func=coverage_filtered.out | tail -1 | grep -o '[0-9]*\.[0-9]*%' | sed 's/%//'); \
	threshold=$(THRESHOLD); \
	echo "Current coverage: $$coverage%"; \
	echo "Threshold: $$threshold%"; \
	if [ $$(echo "$$coverage < $$threshold" | bc -l) -eq 1 ]; then \
		echo "❌ Coverage $$coverage% is below threshold $$threshold%"; \
		exit 1; \
	else \
		echo "✅ Coverage $$coverage% meets threshold $$threshold%"; \
	fi

test-unit-coverage-html:
	@echo "Generating HTML coverage report..."
	@if ! go test -coverprofile=coverage.out $(shell go list ./... | grep -v /tests | grep -v /scripts | grep -v /common/vip180/contracts); then \
		echo "❌ Tests failed! Cannot generate coverage report."; \
		exit 1; \
	fi
	@echo "Filtering out files not required from coverage report..."
	@grep -v "/_test\.go\|/mock_client\.go\|/main\.go" coverage.out > coverage_filtered.out
	@go tool cover -html=coverage_filtered.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Open coverage.html in your browser to view detailed coverage"

test-e2e:
	@echo "Running e2e tests..."
	@echo "Make sure the mesh server is running in solo mode: make docker-solo-up"
	go test -v ./tests/e2e/...

test-e2e-verbose:
	@echo "Running e2e tests with verbose output..."
	@echo "Make sure the mesh server is running in solo mode: make docker-solo-up"
	go test -v -count=1 ./tests/e2e/...

test-e2e-full:
	@echo "Starting full e2e test cycle..."
	@echo "1. Starting solo mode services..."
	@$(MAKE) docker-solo-up
	@echo "2. Waiting for services to be ready..."
	@timeout=60; \
	while [ $$timeout -gt 0 ]; do \
		if curl -s http://localhost:8080/health > /dev/null 2>&1; then \
			echo "✅ Mesh API server is ready!"; \
			break; \
		fi; \
		echo "⏳ Waiting for server... ($$timeout seconds remaining)"; \
		sleep 2; \
		timeout=$$((timeout-2)); \
	done; \
	if [ $$timeout -le 0 ]; then \
		echo "❌ Timeout waiting for server to start"; \
		$(MAKE) docker-solo-down; \
		exit 1; \
	fi
	@echo "3. Running e2e tests..."
	@bash -c '$(MAKE) test-e2e; \
	test_result=$$?; \
	echo "4. Stopping solo mode services..."; \
	$(MAKE) docker-solo-down; \
	if [ $$test_result -eq 0 ]; then \
		echo "✅ All e2e tests passed!"; \
	else \
		echo "❌ Some e2e tests failed!"; \
	fi; \
	exit $$test_result'

test-e2e-vip180:
	@echo "Running VIP180 e2e test..."
	@echo "Make sure the mesh server is running in solo mode: make docker-solo-up"
	go test -v -run TestVIP180Solo ./tests/e2e/...

test-e2e-call:
	@echo "Running Call service e2e test..."
	@echo "Make sure the mesh server is running in solo mode: make docker-solo-up"
	go test -v -run TestCallService_InspectClausesWithVIP180 ./tests/e2e/...

test-e2e-call-full:
	@echo "Starting full Call service e2e test cycle..."
	@echo "1. Starting solo mode services..."
	@$(MAKE) docker-solo-up
	@echo "2. Waiting for services to be ready..."
	@timeout=60; \
	while [ $$timeout -gt 0 ]; do \
		if curl -s http://localhost:8080/health > /dev/null 2>&1; then \
			echo "✅ Mesh API server is ready!"; \
			break; \
		fi; \
		echo "⏳ Waiting for server... ($$timeout seconds remaining)"; \
		sleep 2; \
		timeout=$$((timeout-2)); \
	done; \
	if [ $$timeout -le 0 ]; then \
		echo "❌ Timeout waiting for server to start"; \
		$(MAKE) docker-solo-down; \
		exit 1; \
	fi
	@echo "3. Running Call service e2e test..."
	@bash -c '$(MAKE) test-e2e-call; \
	test_result=$$?; \
	echo "4. Stopping solo mode services..."; \
	$(MAKE) docker-solo-down; \
	if [ $$test_result -eq 0 ]; then \
		echo "✅ Call service e2e test passed!"; \
	else \
		echo "❌ Call service e2e test failed!"; \
	fi; \
	exit $$test_result'

test-e2e-vip180-full:
	@echo "Starting full VIP180 e2e test cycle..."
	@echo "1. Starting solo mode services..."
	@$(MAKE) docker-solo-up
	@echo "2. Waiting for services to be ready..."
	@timeout=60; \
	while [ $$timeout -gt 0 ]; do \
		if curl -s http://localhost:8080/health > /dev/null 2>&1; then \
			echo "✅ Mesh API server is ready!"; \
			break; \
		fi; \
		echo "⏳ Waiting for server... ($$timeout seconds remaining)"; \
		sleep 2; \
		timeout=$$((timeout-2)); \
	done; \
	if [ $$timeout -le 0 ]; then \
		echo "❌ Timeout waiting for server to start"; \
		$(MAKE) docker-solo-down; \
		exit 1; \
	fi
	@echo "3. Running VIP180 e2e test..."
	@bash -c '$(MAKE) test-e2e-vip180; \
	test_result=$$?; \
	echo "4. Stopping solo mode services..."; \
	$(MAKE) docker-solo-down; \
	if [ $$test_result -eq 0 ]; then \
		echo "✅ VIP180 e2e test passed!"; \
	else \
		echo "❌ VIP180 e2e test failed!"; \
	fi; \
	exit $$test_result'

clean:
	@echo "Cleaning Go build artifacts and cache..."
	go clean -cache
	go clean -modcache
	go clean -testcache
	rm -f mesh-server
	@echo "Clean completed!"

# Docker commands
docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-up-build:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-clean:
	docker compose down --rmi all --volumes --remove-orphans

docker-solo-up:
	NETWORK=solo docker compose up --build -d

docker-solo-down:
	NETWORK=solo docker compose down

docker-solo-logs:
	NETWORK=solo docker compose logs -f

# Mesh CLI validation commands
mesh-cli-build:
	@echo "Building mesh-cli Docker image..."
	docker build -f mesh-cli/Dockerfile -t vechain-mesh-cli:latest .

mesh-cli-check-data:
	@if [ -z "$(ENV)" ]; then \
		echo "❌ Error: ENV parameter is required. Use: make mesh-cli-check-data ENV=solo|test|main"; \
		exit 1; \
	fi
	@if [ ! -f "$(PWD)/mesh-cli/$(ENV)/config.json" ]; then \
		echo "❌ Error: Configuration file not found: mesh-cli/$(ENV)/config.json"; \
		exit 1; \
	fi
	@echo "Starting mesh-cli Data API validation for $(ENV) network..."
	@echo "0. Cleaning previous mesh-cli data..."
	@rm -rf mesh-cli/data
	@echo "1. Starting $(ENV) mode services..."
	@if [ "$(ENV)" = "solo" ]; then \
		$(MAKE) docker-solo-up; \
	else \
		NETWORK=$(ENV) $(MAKE) docker-up; \
	fi
	@echo "2. Waiting for services to be ready..."
	@timeout=60; \
	while [ $$timeout -gt 0 ]; do \
		if curl -s http://localhost:8080/health > /dev/null 2>&1; then \
			echo "✅ Mesh API server is ready!"; \
			break; \
		fi; \
		echo "⏳ Waiting for server... ($$timeout seconds remaining)"; \
		sleep 2; \
		timeout=$$((timeout-2)); \
	done; \
	if [ $$timeout -le 0 ]; then \
		echo "❌ Timeout waiting for server to start"; \
		if [ "$(ENV)" = "solo" ]; then \
			$(MAKE) docker-solo-down; \
		else \
			$(MAKE) docker-down; \
		fi; \
		exit 1; \
	fi
	@echo "3. Running mesh-cli Data API validation..."
	@mkdir -p mesh-cli/data
	@bash -c 'docker run --rm \
		--network mesh_vechain-network \
		-v $(PWD)/mesh-cli/$(ENV):/config:ro \
		-v $(PWD)/mesh-cli/data:/data \
		vechain-mesh-cli:latest \
		check:data --configuration-file /config/config.json; \
		validation_result=$$?; \
		echo "4. Stopping $(ENV) mode services..."; \
		if [ "$(ENV)" = "solo" ]; then \
			$(MAKE) docker-solo-down; \
		else \
			$(MAKE) docker-down; \
		fi; \
		if [ $$validation_result -eq 0 ]; then \
			echo "✅ Data API validation passed!"; \
		else \
			echo "❌ Data API validation failed!"; \
		fi; \
		exit $$validation_result'

mesh-cli-check-construction:
	@if [ -z "$(ENV)" ]; then \
		echo "❌ Error: ENV parameter is required. Use: make mesh-cli-check-construction ENV=solo|test|main"; \
		exit 1; \
	fi
	@if [ ! -f "$(PWD)/mesh-cli/$(ENV)/config.json" ]; then \
		echo "❌ Error: Configuration file not found: mesh-cli/$(ENV)/config.json"; \
		exit 1; \
	fi
	@echo "Starting mesh-cli Construction API validation for $(ENV) network..."
	@echo "0. Cleaning previous mesh-cli data..."
	@rm -rf mesh-cli/data
	@echo "1. Starting $(ENV) mode services..."
	@if [ "$(ENV)" = "solo" ]; then \
		$(MAKE) docker-solo-up; \
	else \
		NETWORK=$(ENV) $(MAKE) docker-up; \
	fi
	@echo "2. Waiting for services to be ready..."
	@timeout=60; \
	while [ $$timeout -gt 0 ]; do \
		if curl -s http://localhost:8080/health > /dev/null 2>&1; then \
			echo "✅ Mesh API server is ready!"; \
			break; \
		fi; \
		echo "⏳ Waiting for server... ($$timeout seconds remaining)"; \
		sleep 2; \
		timeout=$$((timeout-2)); \
	done; \
	if [ $$timeout -le 0 ]; then \
		echo "❌ Timeout waiting for server to start"; \
		if [ "$(ENV)" = "solo" ]; then \
			$(MAKE) docker-solo-down; \
		else \
			$(MAKE) docker-down; \
		fi; \
		exit 1; \
	fi
	@echo "3. Running mesh-cli Construction API validation..."
	@echo "⚠️  Note: Construction API validation may take several minutes"
	@mkdir -p mesh-cli/data
	@bash -c 'docker run --rm \
		--network mesh_vechain-network \
		-v $(PWD)/mesh-cli/$(ENV):/config:ro \
		-v $(PWD)/mesh-cli/data:/data \
		vechain-mesh-cli:latest \
		check:construction --configuration-file /config/config.json; \
		validation_result=$$?; \
		echo "4. Stopping $(ENV) mode services..."; \
		if [ "$(ENV)" = "solo" ]; then \
			$(MAKE) docker-solo-down; \
		else \
			$(MAKE) docker-down; \
		fi; \
		if [ $$validation_result -eq 0 ]; then \
			echo "✅ Construction API validation passed!"; \
		else \
			echo "❌ Construction API validation failed!"; \
		fi; \
		exit $$validation_result'

mesh-cli-check-data-solo:
	@$(MAKE) mesh-cli-check-data ENV=solo

mesh-cli-check-construction-solo:
	@$(MAKE) mesh-cli-check-construction ENV=solo
