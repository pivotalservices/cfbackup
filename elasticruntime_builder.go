package cfbackup

import (
	"fmt"

	"github.com/pivotalservices/cfops/tileregistry"
)

// New -- method to generate an initialized elastic runtime
func (s *ElasticRuntimeBuilder) New(tileSpec tileregistry.TileSpec) (tileregistry.Tile, error) {

	// Get installation
	url := fmt.Sprintf(OpsMgrInstallationSettingsURL, tileSpec.OpsManagerHost)
	fmt.Println("Calling NewOpsManGateway...")
	gateway := NewOpsManGateway(url, tileSpec.OpsManagerUser, tileSpec.OpsManagerPass)
	installation, err := getInstallationSettings(gateway)
	if err != nil {
		fmt.Println("*************** error getting installation settings for tile")
		return nil, fmt.Errorf("error getting installation settings for tile %v, %s", tileSpec, err.Error())
	}
	fmt.Printf("2. installation after being unmarshaled: %v", installation)

	// Get ssh key
	// fmt.Println("******CREATING PARSER")
	config := NewConfigurationParser(*installation)
	// fmt.Println("ABOUT TO GET IAAS")
	iaas, err := config.GetIaaS()
	// fmt.Println("I HAVE IAAS.... I THINK")
	if err != nil {
		fmt.Println("I HAVE AN ERROR: %s", err.Error())
		return nil, err
	}
	fmt.Printf("iaas: %v", iaas)
	sshKey := iaas.SSHPrivateKey

	return NewElasticRuntime(*installation, tileSpec.ArchiveDirectory, sshKey), nil
}

func getInstallationSettings(gateway OpsManagerGateway) (*InstallationSettings, error) {
	installation := &InstallationSettings{}
	fmt.Println("calling GetInstallationSettings...")
	apiResponse := gateway.GetInstallationSettings(installation)
	// fmt.Printf("1. installation after being unmarshaled: %v", installation)
	if apiResponse.IsError() {
		fmt.Println("*************** returning error from getInstallationSettings()")
		return nil, fmt.Errorf("error getting installation settings: %s", apiResponse.Message)
	}

	return installation, nil
}

// NewOpsManGateway returns a new OpsManagerGateway
var NewOpsManGateway = func(url, opsmanUser, opsmanPassword string) OpsManagerGateway {
	fmt.Println("IM IN THE REAL DEAL!!!!")
	return NewOpsManagerGateway(url, opsmanUser, opsmanPassword)
}
