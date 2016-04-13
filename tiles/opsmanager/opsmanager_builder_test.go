package opsmanager_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/cfbackup/tileregistry"
	. "github.com/pivotalservices/cfbackup/tiles/opsmanager"
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
	})
})
