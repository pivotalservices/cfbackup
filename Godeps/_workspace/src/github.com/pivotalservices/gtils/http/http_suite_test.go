package http_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHttp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Http Suite")
}

type transportClientInterface interface {
	Do(*http.Request) (*http.Response, error)
}

type mockClientTransport struct{}

func (s *mockClientTransport) Do(req *http.Request) (res *http.Response, err error) {

	return
}
