package blog

import (
	"log/slog"
)

type CriticalError struct {
	Message string
}

func (err *CriticalError) Error() string {
	return err.Message
}

type ValidationError struct {
	Message string
}

func (err *ValidationError) Error() string {
	return err.Message
}

type AuthenticationError struct {
	Message string
}

func (err *AuthenticationError) Error() string {
	return err.Message
}

func OnCriticalError(err error, logger *slog.Logger) {
	logger.Error(err.Error())
}
