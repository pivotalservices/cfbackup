package cfbackup_test

import (
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	. "github.com/pivotalservices/cfbackup"
	"github.com/pivotalservices/cfops/tileregistry"
)

var _ = Describe("OpsManagerBuilder", func() {
	Describe("given a New() method", func() {
		Context("when called with invalid tileSpec connection credentials", func() {
			var controlTileSpec tileregistry.TileSpec
			BeforeEach(func() {
				controlTileSpec = tileregistry.TileSpec{}
			})

			It("then it should lazy load dial and NOT error", func() {
				_, err := new(OpsManagerBuilder).New(controlTileSpec)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})

		Context("when called with a tileSpec", func() {
			var controlTileSpec tileregistry.TileSpec
			BeforeEach(func() {
				controlTileSpec = tileregistry.TileSpec{}
			})

			It("then it should return an initialized OpsManager as a tileregistry.Tile interface", func() {
				tile, _ := new(OpsManagerBuilder).New(controlTileSpec)
				Ω(tile).Should(BeAssignableToTypeOf(new(OpsManager)))
			})
		})

		Context("when the ops manager being created is targeting a foundation which uses PEM keys for auth", func() {
			var (
				server   *ghttp.Server
				fakeUser = "fakeuser"
				fakePass = "fakepass"
			)

			BeforeEach(func() {
				fileBytes, _ := ioutil.ReadFile("./fixtures/installation-settings-1-6-aws.json")
				server = ghttp.NewTLSServer()
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyBasicAuth(fakeUser, fakePass),
						ghttp.RespondWith(http.StatusOK, string(fileBytes[:])),
					),
				)
			})

			AfterEach(func() {
				server.Close()
			})

			It("then it should return a opsmanager which has an executer which uses a pem", func() {
				opsManager, err := new(OpsManagerBuilder).New(tileregistry.TileSpec{
					OpsManagerHost:   strings.Replace(server.URL(), "https://", "", 1),
					AdminUser:        fakeUser,
					AdminPass:        fakePass,
					OpsManagerUser:   "ubuntu",
					OpsManagerPass:   "xxx",
					ArchiveDirectory: "/tmp",
				})
				Ω(err).ShouldNot(HaveOccurred())
				Ω(opsManager.(*OpsManager).SSHPrivateKey).ShouldNot(BeEmpty())
			})
		})
	})
})
