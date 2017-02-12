package vkc

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	// error codes
	codeAuthFailed      = 5
	codeTooManyRequests = 6
	codeFloodControl    = 9
	codeCaptchaNeeded   = 14
	codeNeedValidation  = 17
)

var (
	ErrTooManyRequests = errors.New("too many requests")
	ErrNeedValidation  = errors.New("token needs validation")
	ErrAuthFailled     = errors.New("authorization failed")
	ErrFloodControl    = errors.New("flood control")
	ErrInvalidJson     = errors.New("invalid json response")
)

type VkError struct {
	Code          int                    `json:"error_code"`
	Message       string                 `json:"error_msg"`
	RequestParams []vkErrorRequestParam  `json:"request_params"`
	Rest          map[string]interface{} `json:"-"`
}

type _VkError VkError

func (v *VkError) UnmarshalJSON(bs []byte) error {
	vkErr := _VkError{}

	var err error

	if err = json.Unmarshal(bs, &vkErr); err == nil {
		*v = VkError(vkErr)
	}

	m := make(map[string]interface{})
	if err = json.Unmarshal(bs, &m); err == nil {
		delete(m, "error_code")
		delete(m, "error_msg")
		delete(m, "request_params")
	}
	v.Rest = m

	return err
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

type ErrCaptchaNeeded struct {
	Method     string
	CaptchaSid string
	CaptchaImg string
}

func (e *ErrCaptchaNeeded) Error() string {
	return fmt.Sprintf("captcha needed when executing method '%s'",
		e.Method)
}

func castError(genericErr *VkError) error {
	switch genericErr.Code {
	case codeTooManyRequests:
		return ErrTooManyRequests
	case codeNeedValidation:
		return ErrNeedValidation
	case codeAuthFailed:
		return ErrAuthFailled
	case codeFloodControl:
		return ErrFloodControl
	case codeCaptchaNeeded:
		return &ErrCaptchaNeeded{
			Method:     genericErr.vkMethod(),
			CaptchaSid: genericErr.Rest["captcha_sid"].(string),
			CaptchaImg: genericErr.Rest["captcha_img"].(string),
		}
	default:
		return genericErr
	}
}
