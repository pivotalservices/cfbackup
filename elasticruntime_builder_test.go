package cfbackup_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/cfbackup"
	"github.com/pivotalservices/cfops/tileregistry"
)

var _ = Describe("ElasticRuntimeBuilder", func() {
	Describe("given a New() method", func() {
		XContext("when called with invalid tileSpec connection credentials", func() {
			var controlTileSpec tileregistry.TileSpec
			BeforeEach(func() {
				controlTileSpec = tileregistry.TileSpec{}
			})

			It("then it should return an error", func() {
				_, err := new(ElasticRuntimeBuilder).New(controlTileSpec)
				Ω(err).Should(HaveOccurred())
			})
		})

		Context("when called with a tileSpec", func() {
			var controlTileSpec tileregistry.TileSpec
			BeforeEach(func() {
				controlTileSpec = tileregistry.TileSpec{}
			})

			It("then it should return an initialized ElasticRuntime as a tileregistry.Tile interface", func() {
				tile, _ := new(ElasticRuntimeBuilder).New(controlTileSpec)
				Ω(tile).Should(BeAssignableToTypeOf(new(ElasticRuntime)))
			})
		})
	})
})
