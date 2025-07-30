package usecase

import (
	"context"
	"testing"

	"token/internal/core/domain"
	"token/internal/core/usecase"
	"token/tests"

	"github.com/stretchr/testify/assert"
)

type validateTokenUseCaseTestSetup struct {
	useCase    *usecase.ValidateTokenUseCase
	mockLogger *tests.MockLogger
	validToken string
}

func setupValidateTokenUseCaseTest(t *testing.T) *validateTokenUseCaseTestSetup {
	mockLogger := &tests.MockLogger{}
	secret := "test-secret-key-for-jwt-signing"

	useCase := usecase.NewValidateTokenUseCase(mockLogger, secret).(*usecase.ValidateTokenUseCase)

	tests.SetupCommonLoggerMocks(mockLogger)
	generateUseCase := usecase.NewGenerateTokenUseCase(mockLogger, secret).(*usecase.GenerateTokenUseCase)
	validToken, _ := generateUseCase.GenerateToken(context.Background(), domain.GenerateTokenRequest{
		Subject:   "test@example.com",
		ExpiresIn: "24h",
	})

	return &validateTokenUseCaseTestSetup{
		useCase:    useCase,
		mockLogger: mockLogger,
		validToken: validToken.Token,
	}
}

func (ts *validateTokenUseCaseTestSetup) setupSuccessMocks() {
	tests.SetupSuccessLoggerMocks(ts.mockLogger)
}

func (ts *validateTokenUseCaseTestSetup) setupWarningMocks() {
	tests.SetupWarningLoggerMocks(ts.mockLogger)
}

func TestValidateTokenUseCase_Success(t *testing.T) {
	ts := setupValidateTokenUseCaseTest(t)

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
	ts := setupValidateTokenUseCaseTest(t)

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
	ts := setupValidateTokenUseCaseTest(t)

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
