package elasticruntime_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/cfbackup/tiles/elasticruntime"
)

var _ = Describe("given a TempFile", func() {
	var tempFile *TempFile
	BeforeEach(func() {
		tempFile, _ = NewTempFile("somefilename")
	})
	Context("when calling Close()", func() {
		var tmpFilename string
		BeforeEach(func() {
			tmpFilename = tempFile.FileRef.Name()
			tempFile.Close()
		})
		It("then it should clean up after itself", func() {
			_, err := os.Stat(tmpFilename)
			os.IsNotExist(err)
			Ω(os.IsNotExist(err)).Should(BeTrue())
		})
	})
})
