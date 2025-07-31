package usecase

import (
	"context"
	"testing"

	"token/internal/core/domain"
	"token/internal/core/usecase"
	"token/tests"

	"github.com/stretchr/testify/assert"
)

type generateTokenUseCaseTestSetup struct {
	useCase    *usecase.GenerateTokenUseCase
	mockLogger *tests.MockLogger
}

func setupGenerateTokenUseCaseTest() *generateTokenUseCaseTestSetup {
	mockLogger := &tests.MockLogger{}
	secret := "test-secret-key-for-jwt-signing"

	useCase := usecase.NewGenerateTokenUseCase(mockLogger, secret).(*usecase.GenerateTokenUseCase)

	return &generateTokenUseCaseTestSetup{
		useCase:    useCase,
		mockLogger: mockLogger,
	}
}

func (ts *generateTokenUseCaseTestSetup) setupSuccessMocks() {
	tests.SetupSuccessLoggerMocks(ts.mockLogger)
}

func (ts *generateTokenUseCaseTestSetup) setupWarningMocks() {
	tests.SetupWarningLoggerMocks(ts.mockLogger)
}

func TestGenerateTokenUseCase_Success(t *testing.T) {
	ts := setupGenerateTokenUseCaseTest()

	t.Run("Valid token generation request", func(t *testing.T) {
		request := domain.GenerateTokenRequest{
			Subject:   "test@example.com",
			ExpiresIn: "24h",
		}

		ts.setupSuccessMocks()

		result, err := ts.useCase.GenerateToken(context.Background(), request)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.Token)
		assert.Contains(t, result.Message, "Token generated successfully")
	})

	t.Run("Another valid token generation request", func(t *testing.T) {
		request := domain.GenerateTokenRequest{
			Subject:   "user@test.com",
			ExpiresIn: "1h",
		}

		ts.setupSuccessMocks()

		result, err := ts.useCase.GenerateToken(context.Background(), request)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.Token)
		assert.Contains(t, result.Message, "Token generated successfully")
	})

	t.Run("Token generation with default expiration", func(t *testing.T) {
		request := domain.GenerateTokenRequest{
			Subject: "default@example.com",
		}

		ts.setupSuccessMocks()

		result, err := ts.useCase.GenerateToken(context.Background(), request)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.Token)
		assert.Contains(t, result.Message, "Token generated successfully")
	})
}

func TestGenerateTokenUseCase_InvalidExpiration(t *testing.T) {
	ts := setupGenerateTokenUseCaseTest()

	t.Run("Invalid expiration format", func(t *testing.T) {
		request := domain.GenerateTokenRequest{
			Subject:   "test@example.com",
			ExpiresIn: "invalid-duration",
		}

		ts.setupWarningMocks()

		result, err := ts.useCase.GenerateToken(context.Background(), request)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.Token)
		assert.Contains(t, result.Message, "Token generated successfully")
	})
}

func TestGenerateTokenUseCase_EmptySubject(t *testing.T) {
	ts := setupGenerateTokenUseCaseTest()

	t.Run("Empty subject", func(t *testing.T) {
		request := domain.GenerateTokenRequest{
			Subject:   "",
			ExpiresIn: "24h",
		}

		ts.setupSuccessMocks()

		result, err := ts.useCase.GenerateToken(context.Background(), request)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.Token)
		assert.Contains(t, result.Message, "Token generated successfully")
	})
}
