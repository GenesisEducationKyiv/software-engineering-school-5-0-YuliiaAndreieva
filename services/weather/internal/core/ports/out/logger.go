package out

import sharedlogger "shared/logger"

type Logger = sharedlogger.Logger

type ProviderLogger interface {
	Log(providerName string, responseBody []byte)
	Close() error
}
