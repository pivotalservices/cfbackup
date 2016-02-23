package cfbackup_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/cfbackup"
)

var _ = Describe("given a Products object", func() {
	Context("When properly initialized", func() {
		checkGetMethods("./fixtures/installation-settings-1-7.json", "cf", "nfs_server", 0, 0)
		checkGetMethods("./fixtures/installation-settings-1-6.json", "cf", "nfs_server", 1, 1)
		checkGetMethods("./fixtures/installation-settings-1-6-default.json", "cf", "nfs_server", 1, 1)
		checkGetMethods("./fixtures/installation-settings-1-5.json", "cf", "nfs_server", 1, 1)
	})
})

func checkGetMethods(fixturePath string, productName string, jobName string, propertyCount int, ipsCount int) {
	Context(fmt.Sprintf("when called with a given %s fixture and productName %s", fixturePath, productName), func() {
		var configParser *ConfigurationParser
		var product Products
		var err error
		BeforeEach(func() {
			configParser = NewConfigurationParser(fixturePath)
			product, err = configParser.FindByProductID(productName)
		})
		Context(fmt.Sprintf("when called with a given jobname %s", jobName), func() {
			Describe(fmt.Sprintf("given a GetIPsByJob() - %s", jobName), func() {
				Context("when called with a valid job name", func() {
					It("then it should return ips contained in the product", func() {
						ips := product.GetIPsByJob(jobName)
						Ω(len(ips)).Should(Equal(ipsCount))
					})

				})
			})
			Describe(fmt.Sprintf("given a GetPropertiesByJob - %s", jobName), func() {
				Context("when called with a valid job name", func() {
					It("then it should return valid properties for the job", func() {
						properties, err := product.GetPropertiesByJob(jobName)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(len(properties)).Should(Equal(propertyCount))
					})
				})
			})
			Describe(fmt.Sprintf("given a GetVMCredentialsByJob - %s", jobName), func() {
				Context("when called with a valid job name", func() {
					It("then it should return vm credentials for the job", func() {
						vmCredentials, err := product.GetVMCredentialsByJob(jobName)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(vmCredentials.UserID).ShouldNot(BeEmpty())
						Ω(vmCredentials.Password + vmCredentials.SSLKey).ShouldNot(BeEmpty())
					})
				})
			})

		})
	})
}
