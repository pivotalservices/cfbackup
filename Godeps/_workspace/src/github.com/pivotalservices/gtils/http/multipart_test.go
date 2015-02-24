package http_test

import (
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/gtils/http"
)

var _ = Describe("Multipart", func() {
	Describe("MultiPartBody", func() {
		var (
			multipartConstructor MultiPartBodyFunc = MultiPartBody
		)
		Context("Construct the multipart body successfully", func() {
			It("Should return nil error", func() {
				fileRef, _ := os.Open("fixtures/installation.json")
				body, contentType, err := multipartConstructor("installation[file]", "installation.json", fileRef, nil)
				Ω(err).Should(BeNil())
				Ω(body).ShouldNot(BeNil())
				Ω(contentType).ShouldNot(BeNil())
			})
		})

		Context("Construct the multipart body failed", func() {
			It("Should return error when file is missing", func() {
				fileRef, _ := os.Open("fixtures/installa.json")
				_, _, err := multipartConstructor("installation[file]", "installation.json", fileRef, nil)
				Ω(err).ShouldNot(BeNil())
			})
		})
	})

	Describe("MultiPartUpload", func() {
		var transportClientFunc func() (client interface {
			Do(*http.Request) (*http.Response, error)
		})

		BeforeEach(func() {
			transportClientFunc = NewTransportClient
			NewTransportClient = func() (client interface {
				Do(*http.Request) (*http.Response, error)
			}) {
				client = new(mockClientTransport)
				return
			}
		})

		AfterEach(func() {
			NewTransportClient = transportClientFunc
		})

		It("should do something", func() {
			Ω("hi").Should(Equal("hi"))
		})
	})
})
