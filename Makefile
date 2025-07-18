.PHONY: test test-unit test-integration test-e2e test-all clean

test-all: test-unit test-integration test-e2e

test-unit:
	@echo "📋 Running Unit Tests..."
	go test -v ./internal/core/service/... -tags=unit -count=1

test-integration:
	@echo "🔗 Running Integration Tests..."
	docker compose -f docker-compose.test.yml up -d postgres-integration mailhog
	@echo "Waiting for services to be ready..."
	@sleep 3
	go test -v ./tests/integration/... -tags=integration
	docker compose -f docker-compose.test.yml down

test-e2e:
	@echo "🌐 Running E2E Tests..."
	docker compose -f docker-compose.test.yml up -d postgres-e2e mailhog
	@echo "Waiting for services to be ready..."
	@sleep 3
	@echo "Installing Playwright dependencies..."
	cd tests && npm install && npx playwright install
	@echo "Running E2E Tests..."
	npm run test:e2e
	docker compose -f docker-compose.test.yml down

clean:
	@echo "🧹 Cleaning up test environment..."
	docker compose -f docker-compose.test.yml down --volumes --remove-orphans
	docker system prune -f 