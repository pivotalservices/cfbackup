package cfbackup_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/cfbackup"
)

var _ = Describe("ConfigurationParser", func() {
	Context("When properly initialized", func() {
		checkFindMethods("./fixtures/installation-settings-1-7.json", []string{"cf", "p-bosh"})
		checkFindMethods("./fixtures/installation-settings-1-6.json", []string{"cf", "p-bosh"})
		checkFindMethods("./fixtures/installation-settings-1-6-default.json", []string{"cf", "p-bosh"})
		checkFindMethods("./fixtures/installation-settings-1-5.json", []string{"cf"})
	})

	Describe("NewConfigurationParser", func() {
		keyConfigParser := NewConfigurationParser("./fixtures/installation-settings-1-6-aws.json")
		passConfigParser := NewConfigurationParser("./fixtures/installation-settings-1-6.json")
		describeGetIaaS(keyConfigParser, passConfigParser)
	})

	Describe("NewConfigurationParserFromReader", func() {
		readerAWS, _ := os.Open("./fixtures/installation-settings-1-6-aws.json")
		defer readerAWS.Close()
		keyConfigParser := NewConfigurationParserFromReader(readerAWS)
		reader, _ := os.Open("./fixtures/installation-settings-1-6.json")
		defer reader.Close()
		passConfigParser := NewConfigurationParserFromReader(reader)
		describeGetIaaS(keyConfigParser, passConfigParser)
	})
	checkGetProducts(3, "./fixtures/installation-settings-1-6-default.json")
	checkGetProducts(7, "./fixtures/installation-settings-1-6-aws.json")

})

func checkFindMethods(fixturePath string, productNames []string) {
	Describe(fmt.Sprintf("given a FindByProductID method %s", fixturePath), func() {
		var configParser *ConfigurationParser
		BeforeEach(func() {
			configParser = NewConfigurationParser(fixturePath)
		})
		for _, p := range productNames {
			var productID = p

			Context(fmt.Sprintf("when called with a valid product id %s", productID), func() {

				It("then it should return a the corresponding product", func() {
					defaultInitializedProduct := Products{}
					product, err := configParser.FindByProductID(productID)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(product).ShouldNot(Equal(defaultInitializedProduct))
				})
			})
		}
		Context("when called with a non-existing product id", func() {
			It("then it should return an empty object error", func() {
				_, err := configParser.FindByProductID("i dont exist")
				Ω(err).Should(HaveOccurred())
			})
		})
	})
	Describe("given a FindJobsByProductID method", func() {
		Context("when called with a valid product id", func() {
			It("then it should return a list of jobs for the corresponding product", func() {

			})
		})
		Context("when called with a non-existing product id", func() {
			It("then it should return an empty object error", func() {

			})
		})
	})
}

func checkGetProducts(expectedCount int, fixturePath string) {
	Describe("given a configuration parser", func() {
		var fixtureProductControlCount = expectedCount
		configParser := NewConfigurationParser(fixturePath)
		Context("when calling GetProducts()", func() {
			It("then it should parse out all of the products in the given installation settings", func() {
				products := configParser.GetProducts()
				Ω(len(products)).Should(Equal(fixtureProductControlCount))
			})

		})

	})
}

func describeGetIaaS(keyConfigParser, passConfigParser *ConfigurationParser) {
	Describe("given a GetIaaS() method", func() {
		var controlKey string
		var configParser *ConfigurationParser

		Context("when the installation.json contains a ssh key", func() {
			BeforeEach(func() {
				controlKey = "-----BEGIN RSA PRIVATE KEY-----\nxxxxxxxxxxxxxxxxxxxxx\n-----END RSA PRIVATE KEY-----\n"
				configParser = keyConfigParser
			})
			It("then we should return a valid iaas object", func() {
				iaas, hasKey := configParser.GetIaaS()
				Ω(hasKey).Should(BeTrue())
				Ω(iaas.SSHPrivateKey).Should(Equal(controlKey))
			})
		})
		Context("when the installation.json does not contain a ssh key", func() {
			BeforeEach(func() {
				configParser = passConfigParser
			})
			It("then it should yield a false haskey", func() {
				_, hasKey := configParser.GetIaaS()
				Ω(hasKey).Should(BeFalse())
			})
		})
	})
}
