package core

import "fmt"

type ErrorCode int

const (
	InvalidKey ErrorCode = iota + 1
	RedisError
	MarshalError
	UnmarshalError
)

type ErrorInfo struct {
	Message string
}

var Errors = map[ErrorCode]ErrorInfo{
	InvalidKey:     {Message: "invalid key"},
	RedisError:     {Message: "unable to interact with redis"},
	MarshalError:   {Message: "unable to marshal value"},
	UnmarshalError: {Message: "unable to unmarshal error"},
}

type Error struct {
	Code    ErrorCode
	Message string
	Cause   error
	Key     string
}

func (e Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("cache error (code: %d, key: %s): %s: %v", e.Code, e.Key, e.Message, e.Cause)
	}
	return fmt.Sprintf("cache error (code: %d, key: %s): %s", e.Code, e.Key, e.Message)
}

func NewError(code ErrorCode, key string, cause error) error {
	if info, exists := Errors[code]; exists {
		return &Error{
			Code:    code,
			Message: info.Message,
			Cause:   cause,
			Key:     key,
		}
	}
	return &Error{
		Code:    code,
		Message: "unknown cache error",
		Cause:   cause,
		Key:     key,
	}
}
