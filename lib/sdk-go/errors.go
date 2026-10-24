package sdk

import "fmt"

type SDKError struct {
	Message    string
	StatusCode int
	ErrorBody  string
}

func (e *SDKError) Error() string {
	if e.ErrorBody != "" {
		return e.ErrorBody
	}
	return e.Message
}

type AuthError struct{ SDKError }
type ClientError struct{ SDKError }
type APIError struct{ SDKError }
type TransportError struct{ SDKError }

func authErr(message, body string) error {
	return &AuthError{SDKError{Message: message, StatusCode: 401, ErrorBody: body}}
}

func clientErr(message, body string) error {
	return &ClientError{SDKError{Message: message, StatusCode: 400, ErrorBody: body}}
}

func apiErr(status int, message, body string) error {
	return &APIError{SDKError{Message: message, StatusCode: status, ErrorBody: body}}
}

func transportErr(message string) error {
	return &TransportError{SDKError{Message: message}}
}

func httpStatusMessage(status int) string {
	return fmt.Sprintf("HTTP %d", status)
}
