package cfbackup_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"io"
	"strings"
	"time"

	"github.com/enaml-ops/enaml/enamlbosh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/pivotalservices/cfbackup"
)

var (
	originalNewDirector                     = NewDirector
	getManifest         bool                = true
	getTaskStatus       bool                = true
	changeJobState      bool                = true
	manifest            io.ReadCloser       = ioutil.NopCloser(strings.NewReader("manifest"))
	ip                  string              = "10.10.10.10"
	username            string              = "test"
	password            string              = "test"
	deploymentName      string              = "deployment"
	ccjobs              CloudControllerJobs = CloudControllerJobs{
		CCJob{Job: "job1", Index: 0},
		CCJob{Job: "job2", Index: 1},
		CCJob{Job: "job3", Index: 0},
	}
	task                    Task
	doneTask                Task = Task{}
	changeJobStateCount     int  = 0
	retrieveTaskStatusCount int  = 0
)

type mockDirector struct{}

func (s *mockDirector) GetInfo() (io.ReadCloser, error) {
	return nil, nil
}

func (s *mockDirector) GetCloudControllerVMSet(name string) (io.ReadCloser, error) {
	return os.Open("fixtures/deployment_vms.json")
}

func (director *mockDirector) GetDeploymentManifest(deploymentName string) (io.ReadCloser, error) {
	if !getManifest {
		return nil, errors.New("")
	}
	return manifest, nil
}

func (director *mockDirector) ChangeJobState(deploymentName, jobName, state string, index int) (int, error) {
	changeJobStateCount++
	if !changeJobState {
		return 0, errors.New("")
	}
	return 1, nil
}

func (director *mockDirector) RetrieveTaskStatus(int) (*Task, error) {
	if !getTaskStatus {
		return nil, errors.New("")
	}
	retrieveTaskStatusCount++
	if retrieveTaskStatusCount%2 == 0 {
		return &Task{State: "processing"}, nil
	}
	return &task, nil
}

var _ = Describe("ToggleCcJob", func() {
	TaskPingFreq = time.Millisecond
	var cloudController *CloudController
	var newDirectorCallCount int

	BeforeEach(func() {
		newDirectorCallCount = 0
		NewDirector = func(ip, username, password string, port int) (Bosh, error) {
			newDirectorCallCount++
			return &mockDirector{}, nil
		}
		var err error
		cloudController, err = NewCloudController(ip, username, password, deploymentName, ccjobs)
		Expect(err).ShouldNot(HaveOccurred())
	})
	AfterEach(func() {
		NewDirector = originalNewDirector
	})
	Describe("Toggle All jobs", func() {
		Context("Change Job State failed", func() {
			BeforeEach(func() {
				changeJobState = false
			})
			It("Should return error", func() {
				err := cloudController.Start()
				Ω(err).ShouldNot(BeNil())
			})
		})
		Context("Toggle successfully", func() {
			BeforeEach(func() {
				changeJobState = true
				changeJobStateCount = 0
				task = Task{State: "done"}
				retrieveTaskStatusCount = 0
			})
			It("Should return nil error", func() {
				err := cloudController.Start()
				Ω(err).Should(BeNil())
			})
			It("Should Call changeJobState 3 times with 3 jobs", func() {
				cloudController.Start()
				Ω(changeJobStateCount).Should(Equal(3))
			})
			It("Should Call retriveTaskStatus 5 times with retries when task is processing", func() {
				cloudController.Start()
				Ω(retrieveTaskStatusCount).Should(Equal(5))
			})

			It("should create a new director for each toggle invocation", func() {
				newDirectorCallCount = 0
				cloudController.Start()
				cloudController.Start()
				cloudController.Start()
				Expect(newDirectorCallCount).To(Equal(3))
			})
		})
		Context("Task status is error", func() {
			BeforeEach(func() {
				changeJobState = true
				task = Task{State: "error"}
			})
			It("Should return error", func() {
				err := cloudController.Start()
				Ω(err).ShouldNot(BeNil())
			})
		})
	})

	Describe("NewDirector", func() {

		const tokenResponse = `{
  "access_token":"abcdef01234567890",
  "token_type":"bearer",
  "refresh_token":"0987654321fedcba",
  "expires_in":3599,
  "scope":"opsman.user uaa.admin scim.read opsman.admin scim.write",
  "jti":"foo"
}`

		const basicAuthBoshInfo = `{"name":"enaml-bosh","uuid":"31631ff9-ac41-4eba-a944-04c820633e7f","version":"1.3232.2.0 (00000000)","user":null,"cpi":"aws_cpi","user_authentication":{"type":"basic","options":{}},"features":{"dns":{"status":false,"extras":{"domain_name":null}},"compiled_package_cache":{"status":false,"extras":{"provider":null}},"snapshots":{"status":false}}}`
		const uaaBoshInfo = `{"name":"enaml-bosh","uuid":"9604f9ae-70bf-4c13-8d4d-69ff7f7f091b","version":"1.3232.2.0 (00000000)","user":null,"cpi":"aws_cpi","user_authentication":{"type":"uaa","options":{"url":"%s"}},"features":{"dns":{"status":false,"extras":{"domain_name":null}},"compiled_package_cache":{"status":false,"extras":{"provider":null}},"snapshots":{"status":false}}}`

		var (
			manifestName        = "my-manifest"
			userControl         = "my-user"
			passControl         = "my-pass"
			controlResponseBody = enamlbosh.BoshTask{
				ID:          1180,
				State:       "processing",
				Description: "run errand acceptance_tests from deployment cf-warden",
				Timestamp:   1447033291,
				User:        "admin",
			}
		)

		Context("when called using basic auth", func() {

			var boshclient Bosh
			var server *ghttp.Server
			var err error
			BeforeEach(func() {
				server = ghttp.NewTLSServer()
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/info"),
						ghttp.RespondWith(http.StatusOK, basicAuthBoshInfo),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyBasicAuth(userControl, passControl),
						ghttp.VerifyRequest("GET", fmt.Sprintf("/deployments/%s", manifestName)),
						ghttp.RespondWithJSONEncoded(http.StatusOK, controlResponseBody),
					),
				)

				u, _ := url.Parse(server.URL())
				host, port, _ := net.SplitHostPort(u.Host)
				host = u.Scheme + "://" + host
				portInt, _ := strconv.Atoi(port)
				boshclient, err = originalNewDirector(host, userControl, passControl, portInt)
				Expect(err).ShouldNot(HaveOccurred())
			})

			AfterEach(func() {
				server.Close()
			})

			It("should return a successful response body", func() {
				body, err := boshclient.GetDeploymentManifest(manifestName)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(body).ShouldNot(BeNil())
			})

			It("should include the token in future requests", func() {
				boshclient.GetDeploymentManifest(manifestName)
				lastReq := server.ReceivedRequests()[len(server.ReceivedRequests())-1]
				_, _, hasBasicAuth := lastReq.BasicAuth()
				Ω(hasBasicAuth).Should(BeTrue())
				Ω(lastReq.Header["Authorization"]).Should(ConsistOf("Basic bXktdXNlcjpteS1wYXNz"))
			})
		})

		Context("When called with UAA Auth", func() {
			var boshclient Bosh
			var server *ghttp.Server
			var err error
			BeforeEach(func() {
				server = ghttp.NewTLSServer()
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/info"),
						ghttp.RespondWith(http.StatusOK, fmt.Sprintf(uaaBoshInfo, server.URL())),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/oauth/token"),
						ghttp.RespondWith(http.StatusOK, tokenResponse, http.Header{
							"Content-Type": []string{"application/json"}}),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("/deployments/%s", manifestName)),
						ghttp.RespondWith(http.StatusOK, "[]"),
					),
				)
				u, _ := url.Parse(server.URL())
				host, port, _ := net.SplitHostPort(u.Host)
				host = u.Scheme + "://" + host
				portInt, _ := strconv.Atoi(port)
				boshclient, err = originalNewDirector(host, userControl, passControl, portInt)
				Expect(err).ShouldNot(HaveOccurred())
			})

			AfterEach(func() {
				server.Close()
			})

			It("should return a successful response body", func() {
				body, err := boshclient.GetDeploymentManifest(manifestName)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(body).ShouldNot(BeNil())
			})

			It("should include the token in future requests", func() {
				boshclient.GetDeploymentManifest(manifestName)
				lastReq := server.ReceivedRequests()[len(server.ReceivedRequests())-1]
				_, _, hasBasicAuth := lastReq.BasicAuth()
				Ω(hasBasicAuth).Should(BeFalse())
				Ω(lastReq.Header["Authorization"]).Should(ConsistOf("Bearer abcdef01234567890"))
			})
		})
	})
})
