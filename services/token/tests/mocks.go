package tests

import (
	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debugf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Warnf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	m.Called(format, args)
}

func SetupCommonLoggerMocks(logger *MockLogger) {
	logger.On("Infof", mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything).Return()
	logger.On("Errorf", mock.Anything, mock.Anything).Return()
	logger.On("Warnf", mock.Anything, mock.Anything).Return()
}

func SetupSuccessLoggerMocks(logger *MockLogger) {
	logger.On("Infof", mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything).Return()
}

func SetupErrorLoggerMocks(logger *MockLogger) {
	logger.On("Infof", mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything).Return()
	logger.On("Errorf", mock.Anything, mock.Anything).Return()
}

func SetupWarningLoggerMocks(logger *MockLogger) {
	logger.On("Infof", mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything).Return()
	logger.On("Warnf", mock.Anything, mock.Anything).Return()
}
