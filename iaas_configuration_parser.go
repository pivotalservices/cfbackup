package cfbackup

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

//NewConfigurationParser - constructor for a ConfigurationParser from a json installationsettings file
func NewConfigurationParser(is InstallationSettings) *ConfigurationParser {
	return &ConfigurationParser{
		installationSettings: is,
	}
}

//NewConfigurationParserFromReader - constructor for a ConfigurationParser from a json installationsettings file
func NewConfigurationParserFromReader(settings io.Reader) *ConfigurationParser {
	is := InstallationSettings{}
	b, _ := ioutil.ReadAll(settings)
	json.Unmarshal(b, &is)

	return &ConfigurationParser{
		installationSettings: is,
	}
}

//GetIaaS - get the iaas elements from the installation settings
func (s *ConfigurationParser) GetIaaS() (config IaaSConfiguration, err error) {
	config = s.installationSettings.Infrastructure.IaaSConfig
	fmt.Println("PARSING IAAS config: %v", config)

	fmt.Println("config.SSHPrivateKey: %s", config.SSHPrivateKey)
	if config.SSHPrivateKey == "" {
		err = ErrNoSSLKeyFound
	}
	return
}
