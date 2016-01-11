package cfbackup_test

import (
	. "github.com/pivotalservices/cfbackup"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	ghttp "github.com/pivotalservices/gtils/http"
)

var _ = Describe("GetUploader", func() {
	Describe("given a backup context", func() {
		Context("when the context is s3", func() {
			It("then we should return a MultiPartUploader", func() {
				bc := BackupContext{IsS3: true}
				uploader := GetUploader(bc)
				Ω(uploader).Should(BeAssignableToTypeOf(ghttp.MultiPartUpload))
			})
		})
		Context("when the context is NOT s3", func() {
			It("then we should return a LargeMultiPartUploader", func() {
				bc := BackupContext{IsS3: false}
				uploader := GetUploader(bc)
				Ω(uploader).Should(BeAssignableToTypeOf(ghttp.LargeMultiPartUpload))
			})
		})
	})

})
