package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type RESTRequest struct {
	HTTPReqest   *http.Request
	SeekableBody io.ReadSeeker
}

type RESTClient struct {
	errHandler ErrorHandler
}

type ErrorResponse struct {
	Code        string
	Description string
}

type ErrorHandlerFunc func(*http.Response) ErrorResponse

// HandleError calls f(r).
func (f ErrorHandlerFunc) HandleError(r *http.Response) ErrorResponse {
	return f(r)
}

type ErrorHandler interface {
	HandleError(*http.Response) ErrorResponse
}

func NewRESTClient() *RESTClient {
	return &RESTClient{}
}

func (client *RESTClient) WithErrorHandler(handler ErrorHandler) *RESTClient {
	client.errHandler = handler
	return client
}

func (client RESTClient) NewRequest(method, path, username string, password string, body io.ReadSeeker) (*RESTRequest, APIResponse) {
	if body != nil {
		body.Seek(0, 0)
	}

	request, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, NewAPIResponseWithError("error building request", err)
	}

	if password != "" {
		data := []byte(username + ":" + password)
		auth := base64.StdEncoding.EncodeToString(data)
		request.Header.Set("Authorization", "Basic "+auth)
	}

	request.Header.Set("accept", "application/json")
	request.Header.Set("content-type", "application/json")

	if body != nil {
		switch v := body.(type) {
		case *os.File:
			fileStats, err := v.Stat()
			if err != nil {
				break
			}
			request.ContentLength = fileStats.Size()
		}
	}

	return &RESTRequest{HTTPReqest: request, SeekableBody: body}, NewSuccessfulAPIResponse()
}

func (client RESTClient) Execute(request *RESTRequest) APIResponse {
	_, apiResponse := client.doExecuteAndHandleError(request)
	return apiResponse
}

func (client RESTClient) ExecuteReturnJSONResponse(request *RESTRequest, response interface{}) (http.Header, APIResponse) {
	bytes, headers, apiResponse := client.ExecuteReturnResponseBytes(request)
	if apiResponse.IsNotSuccessful() {
		return headers, apiResponse
	}

	if apiResponse.StatusCode > 203 || strings.TrimSpace(string(bytes)) == "" {
		return headers, apiResponse
	}

	err := json.Unmarshal(bytes, &response)
	if err != nil {
		apiResponse = NewAPIResponseWithError("invalid JSON response from server", err)
	}
	return headers, apiResponse
}

func (client RESTClient) ExecuteReturnResponseBytes(request *RESTRequest) ([]byte, http.Header, APIResponse) {
	rawResponse, apiResponse := client.doExecuteAndHandleError(request)
	if apiResponse.IsNotSuccessful() {
		return nil, rawResponse.Header, apiResponse
	}

	bytes, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		apiResponse = NewAPIResponseWithError("error processing response", err)
	}

	return bytes, rawResponse.Header, apiResponse
}

func (client RESTClient) doExecuteAndHandleError(request *RESTRequest) (*http.Response, APIResponse) {
	var apiResponse APIResponse

	rawResponse, err := doExecute(request.HTTPReqest)
	if err != nil {
		apiResponse = NewAPIResponseWithError("error processing request", err)
		return rawResponse, apiResponse
	}

	if rawResponse.StatusCode > 299 {
		errorResponse := client.errHandler.HandleError(rawResponse)
		message := fmt.Sprintf(
			"server error, status code: %d, error code: %s, message: %s",
			rawResponse.StatusCode,
			errorResponse.Code,
			errorResponse.Description,
		)
		apiResponse = NewAPIResponse(message, errorResponse.Code, rawResponse.StatusCode)
	} else {
		apiResponse = NewAPIResponseWithStatusCode(rawResponse.StatusCode)
	}
	return rawResponse, apiResponse
}

func doExecute(request *http.Request) (*http.Response, error) {
	httpClient := newHTTPClient()

	return httpClient.Do(request)
}

func newHTTPClient() *http.Client {
	tr := NewRoundTripper()
	return &http.Client{
		Transport: tr,
	}
}
