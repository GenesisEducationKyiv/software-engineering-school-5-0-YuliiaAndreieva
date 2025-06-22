import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  use: {
    baseURL: 'http://localhost:8080',
  },
  webServer: {
    command: 'go run cmd/test-server/main.go',
    url: 'http://localhost:8080',
    reuseExistingServer: true,
    timeout: 120 * 1000,
    cwd: '..',
    env: {
      DB_CONN_STR: 'postgres://test:test@localhost:5434/weather_test?sslmode=disable',
      WEATHER_API_KEY: 'test-api-key',
      SMTP_HOST: 'localhost',
      SMTP_PORT: '1025',
      SMTP_USER: 'test@example.com',
      SMTP_PASS: 'test-password',
      PORT: '8080'
    }
  },
}); 