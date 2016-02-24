package cfbackup_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/cfbackup"
)

var _ = Describe("given a InstallationSettings object", func() {
	Context("When properly initialized", func() {
		checkInstallationSettingsIPMethods("./fixtures/installation-settings-1-7.json", "cf", "nfs_server", 1)
		checkInstallationSettingsIPMethods("./fixtures/installation-settings-1-6.json", "cf", "nfs_server", 1)
		checkInstallationSettingsIPMethods("./fixtures/installation-settings-1-6.json", "cf", "mysql", 1)
		checkInstallationSettingsIPMethods("./fixtures/installation-settings-1-6-default.json", "cf", "nfs_server", 1)
		checkInstallationSettingsIPMethods("./fixtures/installation-settings-1-5.json", "cf", "nfs_server", 1)
		
		checkInstallationSettingsCredentialsMethods("./fixtures/installation-settings-1-7.json", "cf", "nfs_server")
		checkInstallationSettingsCredentialsMethods("./fixtures/installation-settings-1-6.json", "cf", "nfs_server")
		checkInstallationSettingsCredentialsMethods("./fixtures/installation-settings-1-6-default.json", "cf", "nfs_server")
		checkInstallationSettingsCredentialsMethods("./fixtures/installation-settings-1-5.json", "cf", "nfs_server")
		
		checkInstallationSettingsFindMethods("./fixtures/installation-settings-1-7.json", []string{"cf", "p-bosh"})
		checkInstallationSettingsFindMethods("./fixtures/installation-settings-1-6.json", []string{"cf", "p-bosh"})
		checkInstallationSettingsFindMethods("./fixtures/installation-settings-1-6-default.json", []string{"cf", "p-bosh"})
		checkInstallationSettingsFindMethods("./fixtures/installation-settings-1-5.json", []string{"cf", "microbosh"})
		
		checkInstallationSettingsFindMethodsWithInvalidProducts("./fixtures/installation-settings-1-7.json")
		checkInstallationSettingsFindMethodsWithInvalidProducts("./fixtures/installation-settings-1-6.json")
		checkInstallationSettingsFindMethodsWithInvalidProducts("./fixtures/installation-settings-1-6-default.json")
		checkInstallationSettingsFindMethodsWithInvalidProducts("./fixtures/installation-settings-1-5.json")
		
		checkInstallationSettingsPostgresJobs("./fixtures/installation-settings-1-7.json", 0)
		checkInstallationSettingsPostgresJobs("./fixtures/installation-settings-1-6.json", 3)
		checkInstallationSettingsPostgresJobs("./fixtures/installation-settings-1-6-default.json", 0)
		checkInstallationSettingsPostgresJobs("./fixtures/installation-settings-1-5.json", 3)
	})
})

func checkInstallationSettingsIPMethods(fixturePath string, productName string, jobName string, ipsCount int) {
	Context(fmt.Sprintf("when called with a given %s fixture", fixturePath), func() {
		var installationSettings InstallationSettings
		BeforeEach(func() {
			configParser := NewConfigurationParser(fixturePath)
			installationSettings = configParser.InstallationSettings
		})

		Describe(fmt.Sprintf("given a FindIPsByProductAndJob() %s, %s", productName, jobName), func() {
			Context("when called with a productName and jobName", func() {
				It("then it should return ips for the job", func() {
					ips, err := installationSettings.FindIPsByProductAndJob(productName, jobName)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(len(ips)).Should(Equal(ipsCount))
				})
			})
		})
	})
}

func checkInstallationSettingsCredentialsMethods(fixturePath string, productName string, jobName string) {
	Context(fmt.Sprintf("when called with a given %s fixture", fixturePath), func() {
		var installationSettings InstallationSettings
		BeforeEach(func() {
			configParser := NewConfigurationParser(fixturePath)
			installationSettings = configParser.InstallationSettings
		})

		Describe(fmt.Sprintf("given a FindVMCredentialsByProductAndJob() %s, %s", productName, jobName), func() {
			Context("when called with a productName and jobName", func() {
				It("then it should return vmcredentials for the job", func() {
					vmcredentials, err := installationSettings.FindVMCredentialsByProductAndJob(productName, jobName)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(vmcredentials.UserID).ShouldNot(BeEmpty())
					Ω(vmcredentials.Password+vmcredentials.SSLKey).ShouldNot(BeEmpty())
				})
			})
		})
		Describe(fmt.Sprintf("given a FindJobByProductAndJobName() %s, %s", productName, jobName), func() {
			Context("when called with a productName and jobName", func() {
				It("then it should return job for the jobName", func() {
					job, err := installationSettings.FindJobByProductAndJobName(productName, jobName)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(job).ShouldNot(Equal(Jobs{}))
				})
			})
		})
	})
}

func checkInstallationSettingsFindMethods(fixturePath string, productNames []string) {
	Context(fmt.Sprintf("when called with a given %s fixture", fixturePath), func() {
		var installationSettings InstallationSettings
		BeforeEach(func() {
			configParser := NewConfigurationParser(fixturePath)
			installationSettings = configParser.InstallationSettings
		})
		Context(fmt.Sprintf("when called with a list of products %s", productNames), func() {
			for _, p := range productNames {
				var productID = p
				Describe("given a FindByProductID method", func() {
					Context(fmt.Sprintf("when called with a valid product id %s", productID), func() {
						It("then it should return a the corresponding product", func() {
							defaultInitializedProduct := Products{}
							product, err := installationSettings.FindByProductID(productID)
							Ω(err).ShouldNot(HaveOccurred())
							Ω(product).ShouldNot(Equal(defaultInitializedProduct))
						})
					})
				})
				Describe("given a FindJobsByProductID", func() {
					Context("when called with a valid product id", func() {
						It("then it should return a list of jobs for the corresponding product", func() {
							jobs := installationSettings.FindJobsByProductID(productID)
							Ω(len(jobs)).ShouldNot(Equal(0))
						})
					})
				})
			}
		})
	})
}

func checkInstallationSettingsFindMethodsWithInvalidProducts(fixturePath string) {
	Context(fmt.Sprintf("when called with a given %s fixture", fixturePath), func() {
		var installationSettings InstallationSettings
		BeforeEach(func() {
			configParser := NewConfigurationParser(fixturePath)
			installationSettings = configParser.InstallationSettings
		})
		Describe("given a FindByProductID method", func() {
			Context("when called with a non-existing product id", func() {
				It("then it should return an empty object error", func() {
					_, err := installationSettings.FindByProductID("i dont exist")
					Ω(err).Should(HaveOccurred())
				})
			})
		})
		Describe("given a FindJobsByProductID", func() {
			Context("when called with a non-existing product id", func() {
				It("then it should return an empty jobs list", func() {
					jobs := installationSettings.FindJobsByProductID("i don't exist")
					Ω(len(jobs)).Should(Equal(0))
				})
			})
		})
	})
}
func checkInstallationSettingsPostgresJobs(fixturePath string, pgJobCount int) {
	Context(fmt.Sprintf("when called with a given %s fixture", fixturePath), func() {
		var installationSettings InstallationSettings
		BeforeEach(func() {
			configParser := NewConfigurationParser(fixturePath)
			installationSettings = configParser.InstallationSettings
		})
		Describe("given a FindCFPostgresJobs", func() {
			It("then it should return the correct number of jobs", func() {
				jobs := installationSettings.FindCFPostgresJobs()
				Ω(len(jobs)).Should(Equal(pgJobCount))
			})
		})
	})
}