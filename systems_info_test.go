package cfbackup_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/cfbackup"
)

var _ = Describe("SystemsInfo", func() {
	var (
		installationSettingsPre16UpgradeVsphereFile, installationSettingsDefaultUpgradeVsphereFile,
		installationSettingsDefaultUpgradeAwsFile, sshKey string
	)

	Describe("NewSystemsInfo", func() {

		installationSettingsPre16UpgradeVsphereFile = "fixtures/installation-settings-1-6.json"
		installationSettingsDefaultUpgradeVsphereFile = "fixtures/installation-settings-1-6-default.json"
		installationSettingsDefaultUpgradeAwsFile = "fixtures/installation-settings-1-6-aws.json"
		sshKey = "valid key"
		Context("Using a 1.6 vsphere installation settings file from a pre-1.6 postgresql upgrade", func() {
			systemsInfo := NewSystemsInfo(installationSettingsPre16UpgradeVsphereFile, sshKey)
			systemDumps := systemsInfo.SystemDumps

			It("should have a systemDumps with the correct number of postgres dbs", func() {
				Ω(systemDumps[ERConsole]).ShouldNot(BeNil())
				Ω(systemDumps[ERCc]).ShouldNot(BeNil())
				Ω(systemDumps[ERUaa]).ShouldNot(BeNil())
				Ω(len(systemsInfo.PersistentSystems())).Should(Equal(5))
			})
			It("should have a systemDumps with the correct number of other SystemInfos", func() {
				Ω(systemDumps[ERMySQL]).ShouldNot(BeNil())
				Ω(systemDumps[ERDirector]).ShouldNot(BeNil())
				Ω(systemDumps[ERNfs]).ShouldNot(BeNil())
				Ω(len(systemsInfo.PersistentSystems())).Should(Equal(5))
			})
		})

		Context("Using a 1.6 vsphere installation settings file from a default mysql/ha upgrade", func() {
			systemsInfo := NewSystemsInfo(installationSettingsDefaultUpgradeVsphereFile, sshKey)
			systemDumps := systemsInfo.SystemDumps

			It("should have a systemDumps with zero postgres dbs", func() {
				Ω(systemDumps[ERConsole]).Should(BeNil())
				Ω(systemDumps[ERCc]).Should(BeNil())
				Ω(systemDumps[ERUaa]).Should(BeNil())
				Ω(len(systemsInfo.PersistentSystems())).Should(Equal(2))
			})
			It("should have a systemDumps with the correct number of other SystemInfos", func() {
				Ω(systemDumps[ERMySQL]).ShouldNot(BeNil())
				Ω(systemDumps[ERDirector]).ShouldNot(BeNil())
				Ω(systemDumps[ERNfs]).ShouldNot(BeNil())
				Ω(len(systemsInfo.PersistentSystems())).Should(Equal(2))
			})
		})

		Context("Using a 1.6 aws installation settings file from a default upgrade", func() {
			systemsInfo := NewSystemsInfo(installationSettingsDefaultUpgradeAwsFile, sshKey)
			systemDumps := systemsInfo.SystemDumps

			It("should have a systemDumps with the correct number of postgres dbs", func() {
				Ω(systemDumps[ERConsole]).Should(BeNil())
				Ω(systemDumps[ERCc]).Should(BeNil())
				Ω(systemDumps[ERUaa]).Should(BeNil())
				Ω(len(systemsInfo.PersistentSystems())).Should(Equal(2))
			})
			It("should have a systemDumps with the correct number of other SystemInfos", func() {
				Ω(systemDumps[ERMySQL]).ShouldNot(BeNil())
				Ω(systemDumps[ERDirector]).ShouldNot(BeNil())
				Ω(systemDumps[ERNfs]).ShouldNot(BeNil())
				Ω(len(systemsInfo.PersistentSystems())).Should(Equal(2))
			})
		})
	})
})
