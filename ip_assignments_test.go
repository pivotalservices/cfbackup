package cfbackup_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/cfbackup"
)

var _ = Describe("given a IPAssignments object", func() {
	Context("When properly initialized", func() {
		checkIPAssignmentsMethods("./fixtures/installation-settings-1-7.json", "cf-f21eea2dbdb8555f89fb", "nfs_server-9356c51805ff7b093988", "c2a389dd5f225e57fafd", 1)
	})
})

func checkIPAssignmentsMethods(fixturePath string, productGuid string, jobGuid string, azName string, ipsCount int) {
	Context(fmt.Sprintf("when called with a given %s fixture", fixturePath), func() {
		var ipAssignments IPAssignments
		BeforeEach(func() {
			configParser := NewConfigurationParser(fixturePath)
			ipAssignments = configParser.InstallationSettings.IPAssignments
		})
		Describe(fmt.Sprintf("given a FindIPsByProductAndJob() %s, %s, %s", productGuid, jobGuid, azName), func() {
			Context("when called with a productName and jobName", func() {
				It("then it should return ips for the job", func() {
					ips, err := ipAssignments.FindIPsByProductGUIDAndJobGUIDAndAvailabilityZoneGUID(productGuid, jobGuid, azName)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(len(ips)).Should(Equal(ipsCount))
				})
			})
		})
		Describe("given a FindIPsByProductAndJob()", func() {
			Context("when called with an invalid productGUID", func() {
				It("then it should return an error", func() {
                    _, err := ipAssignments.FindIPsByProductGUIDAndJobGUIDAndAvailabilityZoneGUID("invalid productGUID", "", "")
					Ω(err).Should(HaveOccurred())
				})
			})
			Context("when called with a valid productGUID and invalid jobGUID", func() {
				It("then it should return an error", func() {
                    _, err := ipAssignments.FindIPsByProductGUIDAndJobGUIDAndAvailabilityZoneGUID("cf-f21eea2dbdb8555f89fb", "invalid job GUID", "")
					Ω(err).Should(HaveOccurred())
				})
			})
			Context("when called with a valid productGUID, jobGUID and invalid azGUID", func() {
				It("then it should return an error", func() {
                    _, err := ipAssignments.FindIPsByProductGUIDAndJobGUIDAndAvailabilityZoneGUID("cf-f21eea2dbdb8555f89fb", "nfs_server-9356c51805ff7b093988", "invalid AZGUID")
					Ω(err).Should(HaveOccurred())
				})
			})
		})
	})
}
