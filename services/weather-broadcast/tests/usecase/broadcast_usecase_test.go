package usecase

import (
	"context"
	"testing"

	"weather-broadcast/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBroadcastUseCase struct {
	mock.Mock
}

func (m *MockBroadcastUseCase) Broadcast(ctx context.Context, frequency domain.Frequency) error {
	args := m.Called(ctx, frequency)
	return args.Error(0)
}

func TestBroadcastUseCase_Success(t *testing.T) {
	mockUseCase := &MockBroadcastUseCase{}

	t.Run("Successful broadcast", func(t *testing.T) {
		mockUseCase.On("Broadcast", mock.Anything, domain.Daily).Return(nil)

		err := mockUseCase.Broadcast(context.Background(), domain.Daily)

		assert.NoError(t, err)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("Another successful broadcast", func(t *testing.T) {
		mockUseCase.On("Broadcast", mock.Anything, domain.Weekly).Return(nil)

		err := mockUseCase.Broadcast(context.Background(), domain.Weekly)

		assert.NoError(t, err)
		mockUseCase.AssertExpectations(t)
	})
}

func TestBroadcastUseCase_Error(t *testing.T) {
	mockUseCase := &MockBroadcastUseCase{}

	t.Run("Broadcast error", func(t *testing.T) {
		mockUseCase.On("Broadcast", mock.Anything, domain.Daily).Return(assert.AnError)

		err := mockUseCase.Broadcast(context.Background(), domain.Daily)

		assert.Error(t, err)
		mockUseCase.AssertExpectations(t)
	})
}

func TestBroadcastUseCase_DifferentFrequencies(t *testing.T) {
	mockUseCase := &MockBroadcastUseCase{}

	t.Run("Monthly frequency", func(t *testing.T) {
		mockUseCase.On("Broadcast", mock.Anything, domain.Monthly).Return(nil)

		err := mockUseCase.Broadcast(context.Background(), domain.Monthly)

		assert.NoError(t, err)
		mockUseCase.AssertExpectations(t)
	})
}
