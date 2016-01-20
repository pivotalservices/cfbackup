package cfbackup

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	ghttp "github.com/pivotalservices/gtils/http"
)

// OpsManagerGateway is an interface to the OpsManager API
type OpsManagerGateway interface {
	GetInstallationSettings(resp interface{}) ghttp.APIResponse
}

type gatewayImpl struct {
	*ghttp.RESTClient
	baseURL  string
	username string
	password string
}

// NewOpsManagerGateway creates a new instance of a OpsManagerGateway
func NewOpsManagerGateway(url, username, password string) OpsManagerGateway {
	return gatewayImpl{
		ghttp.NewRESTClient().WithErrorHandler(errorHandler()),
		url,
		username,
		password,
	}
}

// GetInstallationSettings retrieves all the installation settings from OpsMan
func (gateway gatewayImpl) GetInstallationSettings(resp interface{}) ghttp.APIResponse {
	request, _ := gateway.NewRequest("GET", gateway.baseURL+"installation_settings", gateway.username, gateway.password, nil)
	_, apiResponse := gateway.ExecuteReturnJSONResponse(request, &resp)
	return apiResponse
}

func errorHandler() ghttp.ErrorHandlerFunc {

	type ccErrorResponse struct {
		Code        int
		Description string
	}

	errorHandler := func(response *http.Response) ghttp.ErrorResponse {
		jsonBytes, _ := ioutil.ReadAll(response.Body)
		response.Body.Close()

		ccResp := ccErrorResponse{}
		json.Unmarshal(jsonBytes, &ccResp)

		code := strconv.Itoa(ccResp.Code)

		return ghttp.ErrorResponse{Code: code, Description: ccResp.Description}
	}

	return errorHandler
}
