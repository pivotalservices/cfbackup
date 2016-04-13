package elasticruntime_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/cfbackup/tileregistry"
	. "github.com/pivotalservices/cfbackup/tiles/elasticruntime"
)

var _ = Describe("ElasticRuntimeBuilder", func() {
	Describe("given a New() method", func() {
		Context("when called with invalid tileSpec connection credentials", func() {
			var controlTileSpec tileregistry.TileSpec
			BeforeEach(func() {
				controlTileSpec = tileregistry.TileSpec{}
			})

			It("then it should return an error", func() {
				_, err := new(ElasticRuntimeBuilder).New(controlTileSpec)
				Ω(err).Should(HaveOccurred())
			})
		})

	})
})
