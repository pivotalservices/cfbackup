package http

import "fmt"

type APIResponse struct {
	Message    string
	ErrorCode  string
	StatusCode int
	isError    bool
	isNotFound bool
}

func NewAPIResponse(message string, errorCode string, statusCode int) (apiResponse APIResponse) {
	return APIResponse{
		Message:    message,
		ErrorCode:  errorCode,
		StatusCode: statusCode,
		isError:    true,
	}
}

func NewAPIResponseWithStatusCode(statusCode int) (apiResponse APIResponse) {
	return APIResponse{
		StatusCode: statusCode,
	}
}

func NewAPIResponseWithMessage(message string, a ...interface{}) (apiResponse APIResponse) {
	return APIResponse{
		Message: fmt.Sprintf(message, a...),
		isError: true,
	}
}

func NewAPIResponseWithError(message string, err error) (apiResponse APIResponse) {
	return APIResponse{
		Message: fmt.Sprintf("%s: %s", message, err.Error()),
		isError: true,
	}
}

func NewNotFoundAPIResponse(message string, a ...interface{}) (apiResponse APIResponse) {
	return APIResponse{
		Message:    fmt.Sprintf(message, a...),
		isNotFound: true,
	}
}

func NewSuccessfulAPIResponse() (apiResponse APIResponse) {
	return APIResponse{}
}

func (apiResponse APIResponse) IsError() bool {
	return apiResponse.isError
}

func (apiResponse APIResponse) IsNotFound() bool {
	return apiResponse.isNotFound
}

func (apiResponse APIResponse) IsSuccessful() bool {
	return !apiResponse.IsNotSuccessful()
}

func (apiResponse APIResponse) IsNotSuccessful() bool {
	return apiResponse.IsError() || apiResponse.IsNotFound()
}
