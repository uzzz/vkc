package vkc

import (
	"errors"
	"fmt"
)

const (
	// error codes
	codeAuthFailed      = 5
	codeTooManyRequests = 6
	codeNeedValidation  = 17
)

var (
	ErrTooManyRequests = errors.New("too many requests")
	ErrNeedValidation  = errors.New("token needs validation")
	ErrAuthFailled     = errors.New("authorization failed")
)

type VkError struct {
	Code          int                   `json:"error_code"`
	Message       string                `json:"error_msg"`
	RequestParams []vkErrorRequestParam `json:"request_params"`
}

func (e *VkError) vkMethod() string {
	for _, param := range e.RequestParams {
		if param.Key == "method" {
			return param.Value
		}
	}

	return ""
}

func (e *VkError) Error() string {
	return fmt.Sprintf("error %d when executing method '%s': %s",
		e.Code, e.vkMethod(), e.Message)
}

func castError(genericErr *VkError) error {
	switch genericErr.Code {
	case codeTooManyRequests:
		return ErrTooManyRequests
	case codeNeedValidation:
		return ErrNeedValidation
	case codeAuthFailed:
		return ErrAuthFailled
	default:
		return genericErr
	}
}
