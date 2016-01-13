package cfbackup_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/cfbackup"
)

var _ = Describe("ConfigurationParser", func() {
	Describe("given a GetIaaS() method", func() {
		Context("when the installation.json contains a ssh key", func() {
			var controlKey string
			var configParser *ConfigurationParser
			BeforeEach(func() {
				controlKey = "this is my ssh private key"
				configParser = NewConfigurationParser("./fixtures/installation-settings-1-6-aws.json")
			})
			It("then we should return a valid iaas object", func() {
				iaas, _ := configParser.GetIaaS()
				Ω(iaas.SSHPrivateKey).Should(Equal(controlKey))
			})
		})
	})
})
