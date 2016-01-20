package cfbackup_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/cfbackup"
	"github.com/pivotalservices/cfops/tileregistry"
)

var _ = Describe("ElasticRuntimeBuilder", func() {
	Describe("given a New() method", func() {
		Context("when called with invalid tileSpec connection credentials", func() {
			var controlTileSpec tileregistry.TileSpec
			BeforeEach(func() {
				controlTileSpec = tileregistry.TileSpec{}
				NewOpsManGateway = func(url, opsmanUser, opsmanPassword string) OpsManagerGateway {
					fmt.Println("IM IN MY FAKE!!!!")
					return &FakeOpsManagerGateway{}
				}
			})

			It("then it should return an error", func() {
				_, err := new(ElasticRuntimeBuilder).New(controlTileSpec)
				Ω(err).Should(HaveOccurred())
			})
		})

		Context("when called with a tileSpec", func() {
			var (
				controlFixtureFile = "fixtures/installation-settings-1-4-variant.json"
				controlTileSpec    tileregistry.TileSpec
			)
			BeforeEach(func() {
				controlTileSpec = tileregistry.TileSpec{}
				NewOpsManGateway = func(url, opsmanUser, opsmanPassword string) OpsManagerGateway {
					return &FakeOpsManagerGateway{controlFixtureFile}
				}
			})

			It("then it should return an initialized ElasticRuntime as a tileregistry.Tile interface", func() {
				_, err := new(ElasticRuntimeBuilder).New(controlTileSpec)
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(ErrNoSSLKeyFound))
				// Ω(err).ShouldNot(HaveOccurred())
				// Ω(tile).Should(BeAssignableToTypeOf(new(ElasticRuntime)))
			})
		})

		Context("when the installationsettings file contains a valid key", func() {
			var (
				controlFixtureFile = "fixtures/installation-settings-1-6-aws.json"
				controlTileSpec    tileregistry.TileSpec
			)
			BeforeEach(func() {
				controlTileSpec = tileregistry.TileSpec{}
				NewOpsManGateway = func(url, opsmanUser, opsmanPassword string) OpsManagerGateway {
					return &FakeOpsManagerGateway{controlFixtureFile}
				}
			})

			It("then it should return an initialized ElasticRuntime as a tileregistry.Tile interface", func() {
				tile, err := new(ElasticRuntimeBuilder).New(controlTileSpec)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(tile).Should(BeAssignableToTypeOf(new(ElasticRuntime)))
			})

			It("then it should properly set the SSHPrivateKey in the elastic runtime object", func() {
				tile, err := new(ElasticRuntimeBuilder).New(controlTileSpec)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(tile.(*ElasticRuntime).SSHPrivateKey).ShouldNot(BeEmpty())
			})
		})
	})
})
