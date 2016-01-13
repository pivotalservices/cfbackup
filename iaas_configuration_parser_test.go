package cfbackup_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/cfbackup"
)

var _ = Describe("ConfigurationParser", func() {
	Describe("given a GetIaaS() method", func() {
		var controlKey string
		var configParser *ConfigurationParser

		Context("when the installation.json contains a ssh key", func() {
			BeforeEach(func() {
				controlKey = "-----BEGIN RSA PRIVATE KEY-----\nxxxxxxxxxxxxxxxxxxxxx\n-----END RSA PRIVATE KEY-----\n"
				configParser = NewConfigurationParser("./fixtures/installation-settings-1-6-aws.json")
			})
			It("then we should return a valid iaas object", func() {
				iaas, err := configParser.GetIaaS()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(iaas.SSHPrivateKey).Should(Equal(controlKey))
			})
		})
		Context("when the installation.json does not contain a ssh key", func() {
			BeforeEach(func() {
				configParser = NewConfigurationParser("./fixtures/installation-settings-1-6.json")
			})
			It("then it should return a no-key found error", func() {
				_, err := configParser.GetIaaS()
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(ErrNoSSLKeyFound))
			})
		})
	})
})
