package cfbackup

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/xchapter7x/lo"
)

//NewConfigurationParser - constructor for a ConfigurationParser from a json installationsettings file
func NewConfigurationParser(installationFilePath string) *ConfigurationParser {
	is := InstallationSettings{}
	b, _ := ioutil.ReadFile(installationFilePath)
	if err := json.Unmarshal(b, &is); err != nil {
		lo.G.Error("Unmarshal installation settings error", err)
	}

	return &ConfigurationParser{
		InstallationSettings: is,
	}
}

//NewConfigurationParserFromReader - constructor for a ConfigurationParser from a json installationsettings file
func NewConfigurationParserFromReader(settings io.Reader) *ConfigurationParser {
	is := InstallationSettings{}
	b, _ := ioutil.ReadAll(settings)
	if err := json.Unmarshal(b, &is); err != nil {
		lo.G.Error("Unmarshal installation settings error", err)
	}

	return &ConfigurationParser{
		InstallationSettings: is,
	}
}

//GetIaaS - get the iaas elements from the installation settings
func (s *ConfigurationParser) GetIaaS() (config IaaSConfiguration, hasSSHKey bool) {
	config = s.InstallationSettings.Infrastructure.IaaSConfig
	hasSSHKey = false

	if config.SSHPrivateKey != "" {
		hasSSHKey = true
	}
	return
}

// FindJobsByProductID finds all the jobs in an installation by product id
func (s *ConfigurationParser) FindJobsByProductID(id string) ([]Jobs) {
    cfJobs := []Jobs{}

	for _, product := range s.GetProducts() {
		identifier := product.Identifier
		if identifier == id {
			for _, job := range product.Jobs {
				cfJobs = append(cfJobs, job)
			}
		}
	}
	return cfJobs
}

// FindByProductID finds a product by product id
func (s *ConfigurationParser) FindByProductID(id string) (productResponse Products, err error) {
	var found bool
	for _, product := range s.GetProducts() {
		identifier := product.Identifier
		if identifier == id {
			productResponse = product
			found = true
			break
		}
	}
	if !found {
		err = fmt.Errorf("Product not found %s", id)
	}

	return
}

// FindCFPostgresJobs finds all the postgres jobs in the cf product
func (s *ConfigurationParser) FindCFPostgresJobs() (jobs []Jobs) {

	jobsList := s.FindJobsByProductID("cf")
	for _, job := range jobsList {
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
	return s.InstallationSettings.Products
}
