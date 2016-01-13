package cfbackup

import (
	"encoding/json"
	"io/ioutil"
)

//NewConfigurationParser - constructor for a ConfigurationParser from a json installationsettings file
func NewConfigurationParser(installationFilePath string) *ConfigurationParser {
	is := InstallationSettings{}
	b, _ := ioutil.ReadFile(installationFilePath)
	json.Unmarshal(b, &is)

	return &ConfigurationParser{
		installationSettings: is,
	}
}

//GetIaaS - get the iaas elements from the installation settings
func (s *ConfigurationParser) GetIaaS() (config IaaSConfiguration, err error) {
	config = s.installationSettings.Infrastructure.IaaSConfig
	return
}

type (
	//InstallationSettings - an object to house installationsettings elements from the json
	InstallationSettings struct {
		Infrastructure Infrastructure
	}
	//Infrastructure - a struct to house Infrastructure block elements from the json
	Infrastructure struct {
		IaaSConfig IaaSConfiguration `json:"iaas_configuration"`
	}
	//IaaSConfiguration - a struct to house the IaaSConfiguration block elements from the json
	IaaSConfiguration struct {
		SSHPrivateKey string `json:"ssh_private_key"`
	}
	//ConfigurationParser - the parser to handle installation settings file parsing
	ConfigurationParser struct {
		installationSettings InstallationSettings
	}
)
