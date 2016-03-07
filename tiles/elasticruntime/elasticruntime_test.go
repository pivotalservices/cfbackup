package elasticruntime_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/pivotalservices/cfbackup"
	"github.com/pivotalservices/cfbackup/fakes"
	. "github.com/pivotalservices/cfbackup/tiles/elasticruntime"
	"github.com/pivotalservices/gtils/osutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	ErrorImport = errors.New("failed import")
	ErrorDump   = errors.New("failed dump")
)

type DBInfoMock struct {
	cfbackup.SystemInfo
	ErrState   error
	failImport bool
	failDump   bool
}

func (s *DBInfoMock) GetPersistanceBackup() (dumper cfbackup.PersistanceBackup, err error) {
	dumper = &mockDumper{
		failImport: s.failImport,
		failDump:   s.failDump,
	}
	return
}

func (s *DBInfoMock) Error() error {
	if s.ErrState == nil {
		s.ErrState = s.SystemInfo.Error()
	}
	return s.ErrState
}

type mockDumper struct {
	failImport bool
	failDump   bool
}

func (s mockDumper) Dump(i io.Writer) (err error) {
	i.Write([]byte("sometext"))

	if s.failDump {
		err = ErrorDump
	}
	return
}

func (s mockDumper) Import(i io.Reader) (err error) {
	i.Read([]byte("sometext"))

	if s.failImport {
		err = ErrorImport
	}
	return
}

var _ = Describe("ElasticRuntime", func() {
	Describe("Elasic Runtime legacy (pre-1.6)", func() {
		Describe("Elastic Runtime v1.4 file variant with getpassword IP index error", func() {
			var installationSettingsFilePath = "../../fixtures/installation-settings-1-4-variant.json"
			testERWithVersionSpecificFile(installationSettingsFilePath, "microbosh")
			testPostgresDBBackups(installationSettingsFilePath)
			testMySQLDBBackups(installationSettingsFilePath)
		})

		Describe("ElasticRuntime Version 1.4", func() {
			var installationSettingsFilePath = "../../fixtures/installation-settings-1-4.json"
			testERWithVersionSpecificFile(installationSettingsFilePath, "microbosh")
			testPostgresDBBackups(installationSettingsFilePath)
			testMySQLDBBackups(installationSettingsFilePath)
		})

		Describe("ElasticRuntime Version 1.5", func() {
			var installationSettingsFilePath = "../../fixtures/installation-settings-1-5.json"
			testERWithVersionSpecificFile(installationSettingsFilePath, "microbosh")
			testPostgresDBBackups(installationSettingsFilePath)
			testMySQLDBBackups(installationSettingsFilePath)
		})
	})
	Describe("Elasic Runtime (1.6 and beyond)", func() {
		Describe("ElasticRuntime Version 1.6", func() {
			var installationSettingsFilePath = "../../fixtures/installation-settings-1-6.json"
			testERWithVersionSpecificFile(installationSettingsFilePath, "p-bosh")
			testPostgresDBBackups(installationSettingsFilePath)
			testMySQLDBBackups(installationSettingsFilePath)
		})
		Describe("ElasticRuntime Version 1.7", func() {
			var installationSettingsFilePath = "../../fixtures/installation-settings-1-7.json"
			testERWithVersionSpecificFile(installationSettingsFilePath, "p-bosh")
			testMySQLDBBackups(installationSettingsFilePath)
		})
	})

})

func testMySQLDBBackups(installationSettingsFilePath string) {
	Describe("Mysql Backup / Restore", func() {
		Context("with a valid product and component for mysql", func() {
			var (
				product    = "cf"
				component  = "mysql"
				identifier = "mysql_admin_credentials"
				target     string
				er         ElasticRuntime
				info       = cfbackup.SystemsInfo{
					SystemDumps: map[string]cfbackup.SystemDump{
						"MySqlDBInfo": &DBInfoMock{
							SystemInfo: cfbackup.SystemInfo{
								Product:    product,
								Component:  component,
								Identifier: identifier,
							},
						},
					},
				}
			)

			BeforeEach(func() {
				target, _ = ioutil.TempDir("/tmp", "spec")
				er = ElasticRuntime{
					JSONFile:      installationSettingsFilePath,
					HTTPGateway:   &fakes.MockHTTPGateway{},
					BackupContext: cfbackup.NewBackupContext(target, cfenv.CurrentEnv(), ""),
					SystemsInfo:   info,
				}
				er.ReadAllUserCredentials()
			})

			AfterEach(func() {
				os.Remove(target)
			})

			Context("Backup", func() {
				It("Should write the dumped output to a file in the databaseDir", func() {
					er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["MySqlDBInfo"]}, cfbackup.ExportArchive)
					filename := fmt.Sprintf("%s.backup", component)
					exists, _ := osutils.Exists(path.Join(target, filename))
					Ω(exists).Should(BeTrue())
				})

				It("Should have a nil error and not panic", func() {
					var err error
					Ω(func() {
						err = er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["MySqlDBInfo"]}, cfbackup.ExportArchive)
					}).ShouldNot(Panic())
					Ω(err).Should(BeNil())
				})
			})

			Context("Restore", func() {
				It("should return error if local file does not exist", func() {
					err := er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["MySqlDBInfo"]}, cfbackup.ImportArchive)
					Ω(err).ShouldNot(BeNil())
					Ω(err).Should(BeAssignableToTypeOf(ErrERInvalidPath))
				})

				Context("local file exists", func() {
					var filename = fmt.Sprintf("%s.backup", component)

					BeforeEach(func() {
						file, _ := os.Create(path.Join(target, filename))
						file.Close()
					})

					AfterEach(func() {
						os.Remove(path.Join(target, filename))
					})

					It("should upload file to remote w/o error", func() {
						err := er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["MySqlDBInfo"]}, cfbackup.ImportArchive)
						Ω(err).Should(BeNil())
					})

					Context("write failure", func() {
						var origInfo map[string]cfbackup.SystemDump

						BeforeEach(func() {
							origInfo = info.SystemDumps
							info = cfbackup.SystemsInfo{
								SystemDumps: map[string]cfbackup.SystemDump{
									"MySqlDBInfo": &DBInfoMock{
										failImport: true,
										SystemInfo: cfbackup.SystemInfo{
											Product:    product,
											Component:  component,
											Identifier: identifier,
										},
									},
								},
							}
						})

						AfterEach(func() {
							info.SystemDumps = origInfo
						})
						It("should return error", func() {
							err := er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["MySqlDBInfo"]}, cfbackup.ImportArchive)
							Ω(err).ShouldNot(BeNil())
							Ω(err).ShouldNot(Equal(ErrorImport))
						})
					})
				})
			})
		})
	})
}

func testPostgresDBBackups(installationSettingsFilePath string) {
	Describe("Postgres Backup / Restore", func() {
		Context("with a valid product and component for ccdb", func() {
			var (
				product    = "cf"
				component  = "consoledb"
				identifier = "credentials"
				target     string
				er         ElasticRuntime
				info       = cfbackup.SystemsInfo{
					SystemDumps: map[string]cfbackup.SystemDump{
						"ConsoledbInfo": &DBInfoMock{
							SystemInfo: cfbackup.SystemInfo{
								Product:    product,
								Component:  component,
								Identifier: identifier,
							},
						},
					},
				}
			)

			BeforeEach(func() {
				target, _ = ioutil.TempDir("/tmp", "spec")
				er = ElasticRuntime{
					JSONFile:      installationSettingsFilePath,
					HTTPGateway:   &fakes.MockHTTPGateway{},
					BackupContext: cfbackup.NewBackupContext(target, cfenv.CurrentEnv(), ""),
					SystemsInfo:   info,
				}
				er.ReadAllUserCredentials()
			})

			AfterEach(func() {
				os.Remove(target)
			})

			Context("Backup", func() {
				It("Should write the dumped output to a file in the databaseDir", func() {
					er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["ConsoledbInfo"]}, cfbackup.ExportArchive)
					filename := fmt.Sprintf("%s.backup", component)
					exists, _ := osutils.Exists(path.Join(target, filename))
					Ω(exists).Should(BeTrue())
				})

				It("Should have a nil error and not panic", func() {
					var err error
					Ω(func() {
						err = er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["ConsoledbInfo"]}, cfbackup.ExportArchive)
					}).ShouldNot(Panic())
					Ω(err).Should(BeNil())
				})
			})

			Context("Restore", func() {
				It("should return error if local file does not exist", func() {
					err := er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["ConsoledbInfo"]}, cfbackup.ImportArchive)
					Ω(err).ShouldNot(BeNil())
					Ω(err).Should(BeAssignableToTypeOf(ErrERInvalidPath))
				})

				Context("local file exists", func() {
					var filename = fmt.Sprintf("%s.backup", component)

					BeforeEach(func() {
						file, _ := os.Create(path.Join(target, filename))
						file.Close()
					})

					AfterEach(func() {
						os.Remove(path.Join(target, filename))
					})

					It("should upload file to remote w/o error", func() {
						err := er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["ConsoledbInfo"]}, cfbackup.ImportArchive)
						Ω(err).Should(BeNil())
					})

					Context("write failure", func() {
						var origInfo map[string]cfbackup.SystemDump

						BeforeEach(func() {
							origInfo = info.SystemDumps
							info = cfbackup.SystemsInfo{
								SystemDumps: map[string]cfbackup.SystemDump{
									"ConsoledbInfo": &DBInfoMock{
										failImport: true,
										SystemInfo: cfbackup.SystemInfo{
											Product:    product,
											Component:  component,
											Identifier: identifier,
										},
									},
								},
							}
						})

						AfterEach(func() {
							info.SystemDumps = origInfo
						})
						It("should return error", func() {
							err := er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["ConsoledbInfo"]}, cfbackup.ImportArchive)
							Ω(err).ShouldNot(BeNil())
							Ω(err).ShouldNot(Equal(ErrorImport))
						})
					})
				})
			})
		})

		Context("with a valid product and component for consoledb", func() {
			var (
				product    = "cf"
				component  = "consoledb"
				identifier = "credentials"
				target     string
				er         ElasticRuntime
				info       = cfbackup.SystemsInfo{
					SystemDumps: map[string]cfbackup.SystemDump{
						"ConsoledbInfo": &DBInfoMock{
							SystemInfo: cfbackup.SystemInfo{
								Product:    product,
								Component:  component,
								Identifier: identifier,
							},
						},
					},
				}
			)

			BeforeEach(func() {
				target, _ = ioutil.TempDir("/tmp", "spec")
				er = ElasticRuntime{
					JSONFile:      installationSettingsFilePath,
					HTTPGateway:   &fakes.MockHTTPGateway{},
					BackupContext: cfbackup.NewBackupContext(target, cfenv.CurrentEnv(), ""),
					SystemsInfo:   info,
				}
				er.ReadAllUserCredentials()
			})

			AfterEach(func() {
				os.Remove(target)
			})

			Context("Backup", func() {

				It("Should write the dumped output to a file in the databaseDir", func() {
					er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["ConsoledbInfo"]}, cfbackup.ExportArchive)
					filename := fmt.Sprintf("%s.backup", component)
					exists, _ := osutils.Exists(path.Join(target, filename))
					Ω(exists).Should(BeTrue())
				})

				It("Should have a nil error and not panic", func() {
					var err error
					Ω(func() {
						err = er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["ConsoledbInfo"]}, cfbackup.ExportArchive)
					}).ShouldNot(Panic())
					Ω(err).Should(BeNil())
				})
			})
		})

		Context("with a valid product and component for uaadb", func() {
			var (
				product    = "cf"
				component  = "uaadb"
				identifier = "credentials"
				target     string
				er         ElasticRuntime
				info       = cfbackup.SystemsInfo{
					SystemDumps: map[string]cfbackup.SystemDump{
						"UaadbInfo": &DBInfoMock{
							SystemInfo: cfbackup.SystemInfo{
								Product:    product,
								Component:  component,
								Identifier: identifier,
							},
						},
					},
				}
			)

			BeforeEach(func() {
				target, _ = ioutil.TempDir("/tmp", "spec")
				er = ElasticRuntime{
					JSONFile:      installationSettingsFilePath,
					HTTPGateway:   &fakes.MockHTTPGateway{},
					BackupContext: cfbackup.NewBackupContext(target, cfenv.CurrentEnv(), ""),
					SystemsInfo:   info,
				}
				er.ReadAllUserCredentials()
			})

			AfterEach(func() {
				os.Remove(target)
			})

			Context("Backup", func() {

				It("Should write the dumped output to a file in the databaseDir", func() {
					er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["UaadbInfo"]}, cfbackup.ExportArchive)
					filename := fmt.Sprintf("%s.backup", component)
					exists, _ := osutils.Exists(path.Join(target, filename))
					Ω(exists).Should(BeTrue())
				})

				It("Should have a nil error and not panic", func() {
					var err error
					Ω(func() {
						err = er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["UaadbInfo"]}, cfbackup.ExportArchive)
					}).ShouldNot(Panic())
					Ω(err).Should(BeNil())
				})
			})
		})
	})
}

func testERWithVersionSpecificFile(installationSettingsFilePath string, boshName string) {
	var elasticRuntime *ElasticRuntime
	Describe("NewElasticRuntime", func() {
		BeforeEach(func() {
			elasticRuntime = NewElasticRuntime(installationSettingsFilePath, "", "", "")
		})
		Context("with valid installationSettings file", func() {
			It("ReadAllUserCredentials should return nil error", func() {
				err := elasticRuntime.ReadAllUserCredentials()
				Ω(err).Should(BeNil())
			})
			It("No PersistentSystems should return nil error", func() {
				err := elasticRuntime.ReadAllUserCredentials()
				Ω(err).Should(BeNil())
				for _, psystem := range elasticRuntime.PersistentSystems {
					err = psystem.Error()
					Ω(err).Should(BeNil())
				}
			})
		})
	})
	Describe("Backup / Restore", func() {

		oldNewDirector := cfbackup.NewDirector

		BeforeEach(func() {

			oldNewDirector = cfbackup.NewDirector
			cfbackup.NewDirector = fakes.NewFakeDirector
		})

		AfterEach(func() {
			cfbackup.NewDirector = oldNewDirector
		})
		Context("with valid properties (DirectorInfo)", func() {
			var (
				target string
				er     ElasticRuntime
				info   = cfbackup.SystemsInfo{
					SystemDumps: map[string]cfbackup.SystemDump{
						"DirectorInfo": &cfbackup.SystemInfo{
							Product:    boshName,
							Component:  "director",
							Identifier: "director_credentials",
						},
						"MySqldbInfo": &DBInfoMock{
							SystemInfo: cfbackup.SystemInfo{
								Product:    "cf",
								Component:  "mysql",
								Identifier: "mysql_admin_credentials",
							},
						},
					},
				}
				ps = []cfbackup.SystemDump{info.SystemDumps["MySqldbInfo"]}
			)

			BeforeEach(func() {
				target, _ = ioutil.TempDir("/tmp", "spec")
				er = ElasticRuntime{
					JSONFile:          installationSettingsFilePath,
					HTTPGateway:       &fakes.MockHTTPGateway{},
					BackupContext:     cfbackup.NewBackupContext(target, cfenv.CurrentEnv(), ""),
					SystemsInfo:       info,
					PersistentSystems: ps,
				}
			})

			AfterEach(func() {
				os.Remove(target)
			})

			Context("With valid list of stores", func() {
				Context("Backup", func() {
					It("Should return nil error", func() {
						err := er.Backup()
						Ω(err).Should(BeNil())
					})
				})

				Context("Restore", func() {
					var filename = fmt.Sprintf("%s.backup", "mysql")

					BeforeEach(func() {
						file, _ := os.Create(path.Join(target, filename))
						file.Close()
					})

					AfterEach(func() {
						os.Remove(path.Join(target, filename))
					})

					It("Should return nil error ", func() {
						err := er.Restore()
						Ω(err).Should(BeNil())
					})
				})
			})

			Context("With empty list of stores", func() {
				var psOrig []cfbackup.SystemDump
				BeforeEach(func() {
					psOrig = ps
					er.PersistentSystems = []cfbackup.SystemDump{}
				})

				AfterEach(func() {
					er.PersistentSystems = psOrig
				})
				Context("Backup", func() {
					It("Should return error on empty list of persistence stores", func() {
						err := er.Backup()
						Ω(err).ShouldNot(BeNil())
						Ω(err).Should(Equal(ErrEREmptyDBList))
					})
				})

				Context("Restore", func() {
					It("Should return error on empty list of persistence stores", func() {
						err := er.Restore()
						Ω(err).ShouldNot(BeNil())
						Ω(err).Should(Equal(ErrEREmptyDBList))
					})
				})
			})

			Context("When their is a failure on one of the systemdumps in the array", func() {
				var psOrig []cfbackup.SystemDump
				BeforeEach(func() {
					psOrig = ps
					er.PersistentSystems = []cfbackup.SystemDump{
						&DBInfoMock{
							SystemInfo: cfbackup.SystemInfo{
								Product:    "",
								Component:  "",
								Identifier: "",
							},
						},
						&DBInfoMock{
							SystemInfo: cfbackup.SystemInfo{
								Product:    "",
								Component:  "",
								Identifier: "",
							},
							failDump: true,
						},
					}
				})

				AfterEach(func() {
					er.PersistentSystems = psOrig
				})
				Context("Backup", func() {
					It("should return error if db backup fails", func() {
						err := er.Backup()
						Ω(err).ShouldNot(BeNil())
						Ω(err).Should(Equal(ErrERDBBackup))
					})
				})

				Context("Restore", func() {
					It("should return error if db backup fails", func() {
						err := er.Backup()
						Ω(err).ShouldNot(BeNil())
						Ω(err).Should(Equal(ErrERDBBackup))
					})
				})
			})

		})

		Context("with invalid properties", func() {
			var (
				product    = "cf"
				component  = "consoledb"
				identifier = "credentials"
				target     string
				er         ElasticRuntime
				info       = cfbackup.SystemsInfo{
					SystemDumps: map[string]cfbackup.SystemDump{
						"ConsoledbInfo": &cfbackup.SystemInfo{
							Product:    product,
							Component:  component,
							Identifier: identifier,
						},
					},
				}
			)

			BeforeEach(func() {
				target, _ = ioutil.TempDir("/tmp", "spec")
				er = ElasticRuntime{
					JSONFile:      installationSettingsFilePath,
					HTTPGateway:   &fakes.MockHTTPGateway{true, 500, `{"state":"notdone"}`},
					BackupContext: cfbackup.NewBackupContext(target, cfenv.CurrentEnv(), ""),
					SystemsInfo:   info,
				}
			})

			AfterEach(func() {
				os.Remove(target)
			})

			Context("Backup", func() {

				It("Should not return nil error", func() {
					err := er.Backup()
					Ω(err).ShouldNot(BeNil())
					Ω(err).Should(Equal(ErrERDirectorCreds))
				})

				It("Should not panic", func() {
					var err error
					Ω(func() {
						err = er.Backup()
					}).ShouldNot(Panic())
				})
			})

			Context("Restore", func() {

				It("Should not return nil error", func() {
					err := er.Restore()
					Ω(err).ShouldNot(BeNil())
					Ω(err).Should(Equal(ErrERDirectorCreds))
				})

				It("Should not panic", func() {
					var err error
					Ω(func() {
						err = er.Restore()
					}).ShouldNot(Panic())
				})
			})
		})
	})

	Describe("RunDbBackups function", func() {

		Context("with an systemDump object in a error state", func() {
			var (
				product    = "cf"
				component  = "uaadb"
				identifier = "credentials"
				target     string
				ErrControl = errors.New("fake systemDump error")
				err        error
				er         ElasticRuntime
				info       = cfbackup.SystemsInfo{
					SystemDumps: map[string]cfbackup.SystemDump{
						"UaadbInfo": &DBInfoMock{
							ErrState: ErrControl,
							SystemInfo: cfbackup.SystemInfo{
								Product:    product,
								Component:  component,
								Identifier: identifier,
							},
						},
					},
				}
			)

			BeforeEach(func() {
				target, _ = ioutil.TempDir("/tmp", "spec")
				er = ElasticRuntime{
					JSONFile:      installationSettingsFilePath,
					HTTPGateway:   &fakes.MockHTTPGateway{},
					BackupContext: cfbackup.NewBackupContext(target, cfenv.CurrentEnv(), ""),
					SystemsInfo:   info,
				}
				er.ReadAllUserCredentials()
				err = er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["UaadbInfo"]}, cfbackup.ExportArchive)

			})

			AfterEach(func() {
				os.Remove(target)
			})

			It("should fail fast and tell us the error", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(ErrControl))
			})
		})

		Context("with a invalid product, username and component", func() {
			var (
				product    = "aaaaaaaa"
				component  = "aaaaaaaa"
				identifier = "aaaaaaaa"
				target     string
				er         ElasticRuntime
				info       = cfbackup.SystemsInfo{
					SystemDumps: map[string]cfbackup.SystemDump{
						"ConsoledbInfo": &cfbackup.SystemInfo{
							Product:    product,
							Component:  component,
							Identifier: identifier,
						},
					},
				}
			)

			BeforeEach(func() {
				target, _ = ioutil.TempDir("/tmp", "spec")
				er = ElasticRuntime{
					JSONFile:      installationSettingsFilePath,
					HTTPGateway:   &fakes.MockHTTPGateway{},
					BackupContext: cfbackup.NewBackupContext(target, cfenv.CurrentEnv(), ""),
					SystemsInfo:   info,
				}
				er.ReadAllUserCredentials()
			})

			AfterEach(func() {
				os.Remove(target)
			})

			Context("Backup", func() {

				It("Should not write the dumped output to a file in the databaseDir", func() {
					er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["ConsoledbInfo"]}, cfbackup.ExportArchive)
					filename := fmt.Sprintf("%s.sql", component)
					exists, _ := osutils.Exists(path.Join(target, filename))
					Ω(exists).ShouldNot(BeTrue())
				})

				It("Should have a non nil error and not panic", func() {
					var err error
					Ω(func() {
						err = er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["ConsoledbInfo"]}, cfbackup.ExportArchive)
					}).ShouldNot(Panic())
					Ω(err).ShouldNot(BeNil())
				})
			})

			Context("Restore", func() {
				It("Should have a non nil error and not panic", func() {
					var err error
					Ω(func() {
						err = er.RunDbAction([]cfbackup.SystemDump{info.SystemDumps["ConsoledbInfo"]}, cfbackup.ImportArchive)
					}).ShouldNot(Panic())
					Ω(err).ShouldNot(BeNil())
				})
			})
		})
	})
}
