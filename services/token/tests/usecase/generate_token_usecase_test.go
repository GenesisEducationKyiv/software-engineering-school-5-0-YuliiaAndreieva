package usecase

import (
	"context"
	"testing"

	"token/internal/core/domain"
	"token/internal/core/usecase"
	"token/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type generateTokenUseCaseTestSetup struct {
	useCase    *usecase.GenerateTokenUseCase
	mockLogger *mocks.Logger
}

func setupGenerateTokenUseCaseTest() *generateTokenUseCaseTestSetup {
	mockLogger := &mocks.Logger{}
	secret := "test-secret-key-for-jwt-signing"

	useCase := usecase.NewGenerateTokenUseCase(mockLogger, secret)
	typedUseCase, ok := useCase.(*usecase.GenerateTokenUseCase)
	if !ok {
		panic("Failed to type assert GenerateTokenUseCase")
	}

	return &generateTokenUseCaseTestSetup{
		useCase:    typedUseCase,
		mockLogger: mockLogger,
	}
}

func (ts *generateTokenUseCaseTestSetup) setupSuccessMocks() {
	ts.mockLogger.On("Infof", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Debugf", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Warnf", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
}

func (ts *generateTokenUseCaseTestSetup) setupWarningMocks() {
	ts.mockLogger.On("Infof", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Debugf", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Warnf", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
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
