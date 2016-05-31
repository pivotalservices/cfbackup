package cfbackup_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/cfbackup"
)

var _ = Describe("given a Jobs object", func() {
	Context("when calling GetInstances on a <=1.6 formatted object", func() {
		var jobs []Instances

		BeforeEach(func() {
			configParser := NewConfigurationParser("fixtures/installation-settings-1-6.json")
			installationSettings := configParser.InstallationSettings
			jobsList := installationSettings.FindJobsByProductID("cf")
			for _, job := range jobsList {
				if job.Identifier == "ccdb" ||
					job.Identifier == "uaadb" ||
					job.Identifier == "consoledb" {
					jobs = append(jobs, job.GetInstances()...)
				}
			}
		})
		It("then it should return the single instance record as an array", func() {
			Ω(len(jobs)).Should(Equal(3))
		})
	})

	Context("when calling GetInstances on a 1.7+ formatted object", func() {
		var jobs []Instances

		BeforeEach(func() {
			configParser := NewConfigurationParser("fixtures/installation-settings-1-7-pgsql.json")
			installationSettings := configParser.InstallationSettings
			jobsList := installationSettings.FindJobsByProductID("cf")
			for _, job := range jobsList {
				if job.Identifier == "ccdb" ||
					job.Identifier == "uaadb" ||
					job.Identifier == "consoledb" {
					jobs = append(jobs, job.GetInstances()...)
				}
			}
		})
		It("then it should return the single instance record as an array", func() {
			Ω(len(jobs)).Should(Equal(3))
		})
	})
})
