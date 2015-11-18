package cfbackup_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

			It("then it should return an error", func() {
				_, err := new(OpsManagerBuilder).New(controlTileSpec)
				Ω(err).Should(HaveOccurred())
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
	})
})
