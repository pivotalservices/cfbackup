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
	"github.com/pivotalservices/gtils/persistence"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	ErrorImport = errors.New("failed import")
	ErrorDump   = errors.New("failed dump")
)

type PgInfoMock struct {
	cfbackup.SystemInfo
	ErrState   error
	failImport bool
	failDump   bool
}

func (s *PgInfoMock) GetPersistanceBackup() (dumper cfbackup.PersistanceBackup, err error) {
	dumper = &mockDumper{
		failImport: s.failImport,
		failDump:   s.failDump,
	}
	return
}

func (s *PgInfoMock) Error() error {
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
	Describe("ElasticRuntime Version 1.3", func() {
		var installationSettingsFilePath = "../../fixtures/installation-settings-1-3.json"
		testERWithVersionSpecificFile(installationSettingsFilePath)
	})

	Describe("Elastic Runtime v1.4 file variant with getpassword IP index error", func() {
		var installationSettingsFilePath = "../../fixtures/installation-settings-1-4-variant.json"
		testERWithVersionSpecificFile(installationSettingsFilePath)
	})

	Describe("ElasticRuntime Version 1.4", func() {
		var installationSettingsFilePath = "../../fixtures/installation-settings-1-4.json"
		testERWithVersionSpecificFile(installationSettingsFilePath)
	})

	Describe("ElasticRuntime Version 1.5", func() {
		var installationSettingsFilePath = "../../fixtures/installation-settings-1-5.json"
		testERWithVersionSpecificFile(installationSettingsFilePath)
	})

	Describe("ElasticRuntime Version 1.6", func() {
		os.Setenv(ERVersionEnvFlag, ERVersion16)
		var installationSettingsFilePath = "../../fixtures/installation-settings-1-6.json"
		testERWithVersionSpecificFile(installationSettingsFilePath)
		os.Setenv(ERVersionEnvFlag, "")
	})
	XDescribe("ElasticRuntime Version 1.7", func() {
		os.Setenv(ERVersionEnvFlag, ERVersion16)
		var installationSettingsFilePath = "../../fixtures/installation-settings-1.7.json"
		testERWithVersionSpecificFile(installationSettingsFilePath)
		os.Setenv(ERVersionEnvFlag, "")
	})

	Describe("Given: a er version feature toggle", func() {
		Context("when toggled to v1.6", func() {
			versionBoshName := "p-bosh"
			versionPGDump := "/var/vcap/packages/postgres-9.4.2/bin/pg_dump"
			versionPGRestore := "/var/vcap/packages/postgres-9.4.2/bin/pg_restore"
			oldNewDirector := cfbackup.NewDirector

			BeforeEach(func() {
				oldNewDirector = cfbackup.NewDirector
				cfbackup.NewDirector = fakes.NewFakeDirector
				os.Setenv(ERVersionEnvFlag, ERVersion16)
				cfbackup.SetPGDumpUtilVersions()
			})

			AfterEach(func() {
				cfbackup.NewDirector = oldNewDirector
				os.Setenv(ERVersionEnvFlag, "")
				cfbackup.SetPGDumpUtilVersions()
			})

			It("then it should know the correct bosh name for this version", func() {
				Ω(cfbackup.BoshName()).Should(Equal(versionBoshName))
			})

			It("then it should target the proper vendored postgres utils", func() {
				Ω(persistence.PGDmpDumpBin).Should(Equal(versionPGDump))
				Ω(persistence.PGDmpRestoreBin).Should(Equal(versionPGRestore))
			})
		})
		Context("when NOT toggled to v1.6", func() {
			versionBoshName := "microbosh"
			versionPGDump := "/var/vcap/packages/postgres/bin/pg_dump"
			versionPGRestore := "/var/vcap/packages/postgres/bin/pg_restore"
			oldNewDirector := cfbackup.NewDirector

			BeforeEach(func() {
				oldNewDirector = cfbackup.NewDirector
				cfbackup.NewDirector = fakes.NewFakeDirector
				os.Setenv(ERVersionEnvFlag, "")
				cfbackup.SetPGDumpUtilVersions()
			})

			AfterEach(func() {
				cfbackup.NewDirector = oldNewDirector
				os.Setenv(ERVersionEnvFlag, "")
				cfbackup.SetPGDumpUtilVersions()
			})

			It("then it should know the correct bosh name for this version", func() {
				Ω(cfbackup.BoshName()).Should(Equal(versionBoshName))
			})

			It("then it should target the proper vendored postgres utils", func() {
				Ω(persistence.PGDmpDumpBin).Should(Equal(versionPGDump))
				Ω(persistence.PGDmpRestoreBin).Should(Equal(versionPGRestore))
			})
		})
	})
})

func testERWithVersionSpecificFile(installationSettingsFilePath string) {

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
				product   = cfbackup.BoshName()
				component = "director"
				username  = "director"
				target    string
				er        ElasticRuntime
				info      = cfbackup.SystemsInfo{
					SystemDumps: map[string]cfbackup.SystemDump{
						"DirectorInfo": &cfbackup.SystemInfo{
							Product:   product,
							Component: component,
							Identity:  username,
						},
						"ConsoledbInfo": &PgInfoMock{
							SystemInfo: cfbackup.SystemInfo{
								Product:   product,
								Component: component,
								Identity:  username,
							},
						},
					},
				}
				ps = []cfbackup.SystemDump{info.SystemDumps["ConsoledbInfo"]}
			)

			BeforeEach(func() {
				target, _ = ioutil.TempDir("/tmp", "spec")
				er = ElasticRuntime{
					JSONFile:          installationSettingsFilePath,
					HTTPGateway:       &fakes.MockHTTPGateway{},
					BackupContext:     cfbackup.NewBackupContext(target, cfenv.CurrentEnv()),
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
					var filename = fmt.Sprintf("%s.backup", component)

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

			Context("When db backup fails", func() {
				var psOrig []cfbackup.SystemDump
				BeforeEach(func() {
					psOrig = ps
					er.PersistentSystems = []cfbackup.SystemDump{
						&PgInfoMock{
							SystemInfo: cfbackup.SystemInfo{
								Product:   product,
								Component: component,
								Identity:  username,
							},
						},
						&PgInfoMock{
							SystemInfo: cfbackup.SystemInfo{
								Product:   product,
								Component: component,
								Identity:  username,
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
				product   = "cf"
				component = "consoledb"
				username  = "root"
				target    string
				er        ElasticRuntime
				info      = cfbackup.SystemsInfo{
					SystemDumps: map[string]cfbackup.SystemDump{
						"ConsoledbInfo": &cfbackup.SystemInfo{
							Product:   product,
							Component: component,
							Identity:  username,
						},
					},
				}
			)

			BeforeEach(func() {
				target, _ = ioutil.TempDir("/tmp", "spec")
				er = ElasticRuntime{
					JSONFile:      installationSettingsFilePath,
					HTTPGateway:   &fakes.MockHTTPGateway{true, 500, `{"state":"notdone"}`},
					BackupContext: cfbackup.NewBackupContext(target, cfenv.CurrentEnv()),
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
		Context("with a valid product and component for ccdb", func() {
			var (
				product   = "cf"
				component = "consoledb"
				username  = "root"
				target    string
				er        ElasticRuntime
				info      = cfbackup.SystemsInfo{
					SystemDumps: map[string]cfbackup.SystemDump{
						"ConsoledbInfo": &PgInfoMock{
							SystemInfo: cfbackup.SystemInfo{
								Product:   product,
								Component: component,
								Identity:  username,
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
					BackupContext: cfbackup.NewBackupContext(target, cfenv.CurrentEnv()),
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
									"ConsoledbInfo": &PgInfoMock{
										failImport: true,
										SystemInfo: cfbackup.SystemInfo{
											Product:   product,
											Component: component,
											Identity:  username,
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
				product   = "cf"
				component = "consoledb"
				username  = "root"
				target    string
				er        ElasticRuntime
				info      = cfbackup.SystemsInfo{
					SystemDumps: map[string]cfbackup.SystemDump{
						"ConsoledbInfo": &PgInfoMock{
							SystemInfo: cfbackup.SystemInfo{
								Product:   product,
								Component: component,
								Identity:  username,
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
					BackupContext: cfbackup.NewBackupContext(target, cfenv.CurrentEnv()),
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
				product   = "cf"
				component = "uaadb"
				username  = "root"
				target    string
				er        ElasticRuntime
				info      = cfbackup.SystemsInfo{
					SystemDumps: map[string]cfbackup.SystemDump{
						"UaadbInfo": &PgInfoMock{
							SystemInfo: cfbackup.SystemInfo{
								Product:   product,
								Component: component,
								Identity:  username,
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
					BackupContext: cfbackup.NewBackupContext(target, cfenv.CurrentEnv()),
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

		Context("with an systemDump object in a error state", func() {
			var (
				product    = "cf"
				component  = "uaadb"
				username   = "root"
				target     string
				ErrControl = errors.New("fake systemDump error")
				err        error
				er         ElasticRuntime
				info       = cfbackup.SystemsInfo{
					SystemDumps: map[string]cfbackup.SystemDump{
						"UaadbInfo": &PgInfoMock{
							ErrState: ErrControl,
							SystemInfo: cfbackup.SystemInfo{
								Product:   product,
								Component: component,
								Identity:  username,
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
					BackupContext: cfbackup.NewBackupContext(target, cfenv.CurrentEnv()),
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
				product   = "aaaaaaaa"
				component = "aaaaaaaa"
				username  = "aaaaaaaa"
				target    string
				er        ElasticRuntime
				info      = cfbackup.SystemsInfo{
					SystemDumps: map[string]cfbackup.SystemDump{
						"ConsoledbInfo": &cfbackup.SystemInfo{
							Product:   product,
							Component: component,
							Identity:  username,
						},
					},
				}
			)

			BeforeEach(func() {
				target, _ = ioutil.TempDir("/tmp", "spec")
				er = ElasticRuntime{
					JSONFile:      installationSettingsFilePath,
					HTTPGateway:   &fakes.MockHTTPGateway{},
					BackupContext: cfbackup.NewBackupContext(target, cfenv.CurrentEnv()),
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
