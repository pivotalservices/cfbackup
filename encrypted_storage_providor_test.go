package cfbackup_test

import (
	"encoding/base64"
	"io"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/cfbackup"
	"github.com/pivotalservices/cfbackup/fakes"
)

var _ = Describe("EncryptedStorageProvider", func() {
	Describe("given a NewEncryptedStorageProvider function", func() {
		Context("when given a valid provider and a key", func() {
			controlProvider := new(fakes.FakeStorageProvider)
			controlKey := "my-fake-key"

			It("then it should return a wrapped provider containing the encryption key", func() {
				encrypted, err := NewEncryptedStorageProvider(controlProvider, controlKey)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(encrypted.EncryptionKey).Should(Equal(controlKey))
			})
		})

		Context("when NOT given a valid key", func() {
			controlProvider := new(fakes.FakeStorageProvider)
			controlInvalidKey := ""

			It("then it should return a wrapped provider containing the encryption key", func() {
				_, err := NewEncryptedStorageProvider(controlProvider, controlInvalidKey)
				Ω(err).Should(HaveOccurred())
			})
		})
	})

	Describe("given an EncryptedStorageProvider object", func() {
		var msp *fakes.MockStringStorageProvider
		var encrpytedProvidor *EncryptedStorageProvider
		var controlEncrpytionKey = "my-fake-encryption-key12"

		BeforeEach(func() {
			msp = fakes.NewMockStringStorageProvider()
			encrpytedProvidor, _ = NewEncryptedStorageProvider(msp, controlEncrpytionKey)
		})

		Describe("given a Reader method", func() {
			Context("when called with valid arguments", func() {
				var reader io.ReadCloser
				var err error
				var controlMessage = "hello there"
				var controlMessageCrypt = "TJ-0JVfGYYCoQyg="

				BeforeEach(func() {
					writer, _ := encrpytedProvidor.Writer("")
					io.WriteString(writer, controlMessage)
					reader, err = encrpytedProvidor.Reader("")
				})
				It("then it should run without error", func() {
					Ω(err).ShouldNot(HaveOccurred())
				})
				It("then it should not allow the underlying reader access to decryption mechanism", func() {
					b, _ := ioutil.ReadAll(msp)
					Ω(msp.String()).ShouldNot(Equal(controlMessage))
					Ω(base64.URLEncoding.EncodeToString(b)).Should(Equal(controlMessageCrypt))
				})
				It("then it should return a reader that de-crypts", func() {
					b, _ := ioutil.ReadAll(reader)
					Ω(msp.String()).ShouldNot(Equal(controlMessage))
					Ω(string(b)).Should(Equal(controlMessage))
				})
			})
		})

		Describe("given a Writer method", func() {
			Context("when called with valid arguments", func() {
				var writer io.WriteCloser
				var err error
				var controlMessage = "hello there"
				var controlMessageCrypt = "TJ-0JVfGYYCoQyg="
				BeforeEach(func() {
					writer, err = encrpytedProvidor.Writer("")
				})
				It("then it should run without error", func() {
					Ω(err).ShouldNot(HaveOccurred())
				})
				It("then it should return a writer that encrypts", func() {
					io.WriteString(writer, controlMessage)
					Ω(base64.URLEncoding.EncodeToString(msp.Bytes())).Should(Equal(controlMessageCrypt))
				})
			})
		})
	})
})
