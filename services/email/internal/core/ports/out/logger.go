//go:generate mockery --dir . --output ../../../../tests/mocks --outpkg mocks --filename logger_mock.go --structname Logger --name Logger
package out

import sharedlogger "shared/logger"

// Logger interface просто використовує shared logger
type Logger = sharedlogger.Logger
