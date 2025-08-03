//go:generate mockery --dir . --output ../../../../tests/mocks --outpkg mocks --filename logger_mock.go --structname Logger
package out

type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}
