# Test Commands

All commands should be run from the root of the project.

## Scripts (Linux/Mac)

- `./scripts/run-unit-tests.sh`: Runs all unit tests.
- `./scripts/run-integration-tests.sh`: Runs all integration tests.
- `./scripts/run-e2e-tests.sh`: Runs all End-to-End (E2E) tests.
- `./scripts/run-all-tests.sh`: Runs all unit, integration, and E2E tests.

## Manual Execution

### Unit Tests
- `go test -v ./internal/core/service/... -tags=unit -count=1`: Runs all unit tests without Docker.

### Integration Tests
1. `docker compose -f docker-compose.test.yml up -d postgres-integration mailhog`: Starts PostgreSQL and MailHog for integration tests.
2. `go test -v ./tests/integration/... -tags=integration -count=1`: Runs integration test files.
3. `docker compose -f docker-compose.test.yml down`: Stops and removes the test environment.

### E2E Tests
1. `docker compose -f docker-compose.test.yml up -d postgres-e2e mailhog`: Starts PostgreSQL for E2E tests.
2. `cd tests && npm run test:e2e`: Runs Playwright E2E tests.
3. `cd .. && docker compose -f docker-compose.test.yml down`: Stops and removes the test environment. 