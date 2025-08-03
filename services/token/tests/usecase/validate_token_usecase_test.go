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

type validateTokenUseCaseTestSetup struct {
	useCase    *usecase.ValidateTokenUseCase
	mockLogger *mocks.Logger
	validToken string
}

func setupValidateTokenUseCaseTest() *validateTokenUseCaseTestSetup {
	mockLogger := &mocks.Logger{}
	secret := "test-secret-key-for-jwt-signing"

	useCase := usecase.NewValidateTokenUseCase(mockLogger, secret).(*usecase.ValidateTokenUseCase)

	// Налаштовуємо моки для логера
	mockLogger.On("Infof", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warnf", mock.Anything, mock.Anything).Return()
	mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()

	generateUseCase := usecase.NewGenerateTokenUseCase(mockLogger, secret).(*usecase.GenerateTokenUseCase)
	validToken, err := generateUseCase.GenerateToken(context.Background(), domain.GenerateTokenRequest{
		Subject:   "test@example.com",
		ExpiresIn: "24h",
	})
	if err != nil {
		panic("Failed to generate test token")
	}

	return &validateTokenUseCaseTestSetup{
		useCase:    useCase,
		mockLogger: mockLogger,
		validToken: validToken.Token,
	}
}

func (ts *validateTokenUseCaseTestSetup) setupSuccessMocks() {
	ts.mockLogger.On("Infof", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Warnf", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
}

func (ts *validateTokenUseCaseTestSetup) setupWarningMocks() {
	ts.mockLogger.On("Infof", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Warnf", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
}

func TestValidateTokenUseCase_Success(t *testing.T) {
	ts := setupValidateTokenUseCaseTest()

	t.Run("Valid token validation", func(t *testing.T) {
		request := domain.ValidateTokenRequest{
			Token: ts.validToken,
		}

		ts.setupSuccessMocks()

		result, err := ts.useCase.ValidateToken(context.Background(), request)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.True(t, result.Valid)
		assert.Contains(t, result.Message, "Token is valid")
	})
}

func TestValidateTokenUseCase_InvalidToken(t *testing.T) {
	ts := setupValidateTokenUseCaseTest()

	t.Run("Invalid token", func(t *testing.T) {
		request := domain.ValidateTokenRequest{
			Token: "invalid.token.here",
		}

		ts.setupWarningMocks()

		result, err := ts.useCase.ValidateToken(context.Background(), request)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Message, "Token is invalid")
	})

	t.Run("Empty token", func(t *testing.T) {
		request := domain.ValidateTokenRequest{
			Token: "",
		}

		ts.setupWarningMocks()

		result, err := ts.useCase.ValidateToken(context.Background(), request)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Message, "Token is invalid")
	})

	t.Run("Malformed token", func(t *testing.T) {
		request := domain.ValidateTokenRequest{
			Token: "not.a.valid.jwt.token",
		}

		ts.setupWarningMocks()

		result, err := ts.useCase.ValidateToken(context.Background(), request)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Message, "Token is invalid")
	})
}

func TestValidateTokenUseCase_DifferentSecret(t *testing.T) {
	ts := setupValidateTokenUseCaseTest()

	t.Run("Token signed with different secret", func(t *testing.T) {
		differentSecret := "different-secret-key"
		differentUseCase := usecase.NewGenerateTokenUseCase(ts.mockLogger, differentSecret).(*usecase.GenerateTokenUseCase)

		differentToken, _ := differentUseCase.GenerateToken(context.Background(), domain.GenerateTokenRequest{
			Subject:   "test@example.com",
			ExpiresIn: "24h",
		})

		request := domain.ValidateTokenRequest{
			Token: differentToken.Token,
		}

		ts.setupWarningMocks()

		result, err := ts.useCase.ValidateToken(context.Background(), request)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Message, "Token is invalid")
	})
}
