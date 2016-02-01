package cfbackup

import (
	"encoding/json"
	"io"
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
func (s *ConfigurationParser) GetIaaS() (config IaaSConfiguration, hasSSHKey bool) {
	config = s.installationSettings.Infrastructure.IaaSConfig
	hasSSHKey = false

	if config.SSHPrivateKey != "" {
		hasSSHKey = true
	}
	return
}

// FindJobsByProductID finds all the jobs in an installation by product id
func (s *ConfigurationParser) FindJobsByProductID(id string) []Jobs {
	cfJobs := []Jobs{}

	for _, product := range s.GetProducts() {
		identifier := product.Identifer
		if identifier == id {
			for _, job := range product.Jobs {
				cfJobs = append(cfJobs, job)
			}
		}
	}
	return cfJobs
}

// FindCFPostgresJobs finds all the postgres jobs in the cf product
func (s *ConfigurationParser) FindCFPostgresJobs() []Jobs {
	jobs := []Jobs{}

	for _, job := range s.FindJobsByProductID("cf") {
		if isPostgres(job.Identifier, job.Instances) {
			jobs = append(jobs, job)
		}
	}
	return jobs
}

func isPostgres(job string, instances []Instances) bool {
	pgdbs := []string{"ccdb", "uaadb", "consoledb"}

	for _, pgdb := range pgdbs {
		if pgdb == job {
			for _, instances := range instances {
				val := instances.Value
				if val >= 1 {
					return true
				}
			}
		}
	}
	return false
}

//GetProducts - get the products array
func (s *ConfigurationParser) GetProducts() (products []Products) {
	return s.installationSettings.Products
}
