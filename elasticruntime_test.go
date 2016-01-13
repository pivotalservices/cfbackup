package cfbackup_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/cloudfoundry-community/go-cfenv"
	. "github.com/pivotalservices/cfbackup"
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
	SystemInfo
	failImport bool
	failDump   bool
}

func (s *PgInfoMock) GetPersistanceBackup() (dumper PersistanceBackup, err error) {
	dumper = &mockDumper{
		failImport: s.failImport,
		failDump:   s.failDump,
	}
	return
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
		var installationSettingsFilePath = "fixtures/installation-settings-1-3.json"
		testERWithVersionSpecificFile(installationSettingsFilePath)
	})

	Describe("Elastic Runtime v1.4 file variant with getpassword IP index error", func() {
		var installationSettingsFilePath = "fixtures/installation-settings-1-4-variant.json"
		testERWithVersionSpecificFile(installationSettingsFilePath)
	})

	Describe("ElasticRuntime Version 1.4", func() {
		var installationSettingsFilePath = "fixtures/installation-settings-1-4.json"
		testERWithVersionSpecificFile(installationSettingsFilePath)
	})

	Describe("ElasticRuntime Version 1.5", func() {
		var installationSettingsFilePath = "fixtures/installation-settings-1-5.json"
		testERWithVersionSpecificFile(installationSettingsFilePath)
	})

	Describe("ElasticRuntime Version 1.6", func() {
		os.Setenv(ERVersionEnvFlag, ERVersion16)
		var installationSettingsFilePath = "fixtures/installation-settings-1-6.json"
		testERWithVersionSpecificFile(installationSettingsFilePath)
		os.Setenv(ERVersionEnvFlag, "")
	})

	Describe("Given: a er version feature toggle", func() {
		Context("when toggled to v1.6", func() {
			versionBoshName := "p-bosh"
			versionPGDump := "/var/vcap/packages/postgres-9.4.2/bin/pg_dump"
			versionPGRestore := "/var/vcap/packages/postgres-9.4.2/bin/pg_restore"

			BeforeEach(func() {
				os.Setenv(ERVersionEnvFlag, ERVersion16)
				SetPGDumpUtilVersions()
			})

			AfterEach(func() {
				os.Setenv(ERVersionEnvFlag, "")
				SetPGDumpUtilVersions()
			})

			It("then it should know the correct bosh name for this version", func() {
				Ω(BoshName()).Should(Equal(versionBoshName))
			})

			It("then it should target the proper vendored postgres utils", func() {
				Ω(persistence.PGDMP_DUMP_BIN).Should(Equal(versionPGDump))
				Ω(persistence.PGDMP_RESTORE_BIN).Should(Equal(versionPGRestore))
			})
		})
		Context("when NOT toggled to v1.6", func() {
			versionBoshName := "microbosh"
			versionPGDump := "/var/vcap/packages/postgres/bin/pg_dump"
			versionPGRestore := "/var/vcap/packages/postgres/bin/pg_restore"

			BeforeEach(func() {
				os.Setenv(ERVersionEnvFlag, "")
				SetPGDumpUtilVersions()
			})

			AfterEach(func() {
				os.Setenv(ERVersionEnvFlag, "")
				SetPGDumpUtilVersions()
			})

			It("then it should know the correct bosh name for this version", func() {
				Ω(BoshName()).Should(Equal(versionBoshName))
			})

			It("then it should target the proper vendored postgres utils", func() {
				Ω(persistence.PGDMP_DUMP_BIN).Should(Equal(versionPGDump))
				Ω(persistence.PGDMP_RESTORE_BIN).Should(Equal(versionPGRestore))
			})
		})
	})
})

func testERWithVersionSpecificFile(installationSettingsFilePath string) {
	Describe("Backup / Restore", func() {
		Context("with valid properties (DirectorInfo)", func() {
			var (
				product   = BoshName()
				component = "director"
				username  = "director"
				target    string
				er        ElasticRuntime
				info      = map[string]SystemDump{
					"DirectorInfo": &SystemInfo{
						Product:   product,
						Component: component,
						Identity:  username,
					},
					"ConsoledbInfo": &PgInfoMock{
						SystemInfo: SystemInfo{
							Product:   product,
							Component: component,
							Identity:  username,
						},
					},
				}
				ps = []SystemDump{info["ConsoledbInfo"]}
			)

			BeforeEach(func() {
				target, _ = ioutil.TempDir("/tmp", "spec")
				er = ElasticRuntime{
					JSONFile:          installationSettingsFilePath,
					HTTPGateway:       &MockHttpGateway{},
					BackupContext:     NewBackupContext(target, cfenv.CurrentEnv()),
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
				var psOrig []SystemDump
				BeforeEach(func() {
					psOrig = ps
					er.PersistentSystems = []SystemDump{}
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
				var psOrig []SystemDump
				BeforeEach(func() {
					psOrig = ps
					er.PersistentSystems = []SystemDump{
						&PgInfoMock{
							SystemInfo: SystemInfo{
								Product:   product,
								Component: component,
								Identity:  username,
							},
						},
						&PgInfoMock{
							SystemInfo: SystemInfo{
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
				info      = map[string]SystemDump{
					"ConsoledbInfo": &SystemInfo{
						Product:   product,
						Component: component,
						Identity:  username,
					},
				}
			)

			BeforeEach(func() {
				target, _ = ioutil.TempDir("/tmp", "spec")
				er = ElasticRuntime{
					JSONFile:      installationSettingsFilePath,
					HTTPGateway:   &MockHttpGateway{true, 500, `{"state":"notdone"}`},
					BackupContext: NewBackupContext(target, cfenv.CurrentEnv()),
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
				info      = map[string]SystemDump{
					"ConsoledbInfo": &PgInfoMock{
						SystemInfo: SystemInfo{
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
					HTTPGateway:   &MockHttpGateway{},
					BackupContext: NewBackupContext(target, cfenv.CurrentEnv()),
					SystemsInfo:   info,
				}
				er.ReadAllUserCredentials()
			})

			AfterEach(func() {
				os.Remove(target)
			})

			Context("Backup", func() {
				It("Should write the dumped output to a file in the databaseDir", func() {
					er.RunDbAction([]SystemDump{info["ConsoledbInfo"]}, ExportArchive)
					filename := fmt.Sprintf("%s.backup", component)
					exists, _ := osutils.Exists(path.Join(target, filename))
					Ω(exists).Should(BeTrue())
				})

				It("Should have a nil error and not panic", func() {
					var err error
					Ω(func() {
						err = er.RunDbAction([]SystemDump{info["ConsoledbInfo"]}, ExportArchive)
					}).ShouldNot(Panic())
					Ω(err).Should(BeNil())
				})
			})

			Context("Restore", func() {
				It("should return error if local file does not exist", func() {
					err := er.RunDbAction([]SystemDump{info["ConsoledbInfo"]}, ImportArchive)
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
						err := er.RunDbAction([]SystemDump{info["ConsoledbInfo"]}, ImportArchive)
						Ω(err).Should(BeNil())
					})

					Context("write failure", func() {
						var origInfo map[string]SystemDump

						BeforeEach(func() {
							origInfo = info
							info = map[string]SystemDump{
								"ConsoledbInfo": &PgInfoMock{
									failImport: true,
									SystemInfo: SystemInfo{
										Product:   product,
										Component: component,
										Identity:  username,
									},
								},
							}
						})

						AfterEach(func() {
							info = origInfo
						})
						It("should return error", func() {
							err := er.RunDbAction([]SystemDump{info["ConsoledbInfo"]}, ImportArchive)
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
				info      = map[string]SystemDump{
					"ConsoledbInfo": &PgInfoMock{
						SystemInfo: SystemInfo{
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
					HTTPGateway:   &MockHttpGateway{},
					BackupContext: NewBackupContext(target, cfenv.CurrentEnv()),
					SystemsInfo:   info,
				}
				er.ReadAllUserCredentials()
			})

			AfterEach(func() {
				os.Remove(target)
			})

			Context("Backup", func() {

				It("Should write the dumped output to a file in the databaseDir", func() {
					er.RunDbAction([]SystemDump{info["ConsoledbInfo"]}, ExportArchive)
					filename := fmt.Sprintf("%s.backup", component)
					exists, _ := osutils.Exists(path.Join(target, filename))
					Ω(exists).Should(BeTrue())
				})

				It("Should have a nil error and not panic", func() {
					var err error
					Ω(func() {
						err = er.RunDbAction([]SystemDump{info["ConsoledbInfo"]}, ExportArchive)
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
				info      = map[string]SystemDump{
					"UaadbInfo": &PgInfoMock{
						SystemInfo: SystemInfo{
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
					HTTPGateway:   &MockHttpGateway{},
					BackupContext: NewBackupContext(target, cfenv.CurrentEnv()),
					SystemsInfo:   info,
				}
				er.ReadAllUserCredentials()
			})

			AfterEach(func() {
				os.Remove(target)
			})

			Context("Backup", func() {

				It("Should write the dumped output to a file in the databaseDir", func() {
					er.RunDbAction([]SystemDump{info["UaadbInfo"]}, ExportArchive)
					filename := fmt.Sprintf("%s.backup", component)
					exists, _ := osutils.Exists(path.Join(target, filename))
					Ω(exists).Should(BeTrue())
				})

				It("Should have a nil error and not panic", func() {
					var err error
					Ω(func() {
						err = er.RunDbAction([]SystemDump{info["UaadbInfo"]}, ExportArchive)
					}).ShouldNot(Panic())
					Ω(err).Should(BeNil())
				})
			})

			Context("Restore", func() {

			})
		})

		Context("with a invalid product, username and component", func() {
			var (
				product   = "aaaaaaaa"
				component = "aaaaaaaa"
				username  = "aaaaaaaa"
				target    string
				er        ElasticRuntime
				info      = map[string]SystemDump{
					"ConsoledbInfo": &SystemInfo{
						Product:   product,
						Component: component,
						Identity:  username,
					},
				}
			)

			BeforeEach(func() {
				target, _ = ioutil.TempDir("/tmp", "spec")
				er = ElasticRuntime{
					JSONFile:      installationSettingsFilePath,
					HTTPGateway:   &MockHttpGateway{},
					BackupContext: NewBackupContext(target, cfenv.CurrentEnv()),
					SystemsInfo:   info,
				}
				er.ReadAllUserCredentials()
			})

			AfterEach(func() {
				os.Remove(target)
			})

			Context("Backup", func() {

				It("Should not write the dumped output to a file in the databaseDir", func() {
					er.RunDbAction([]SystemDump{info["ConsoledbInfo"]}, ExportArchive)
					filename := fmt.Sprintf("%s.sql", component)
					exists, _ := osutils.Exists(path.Join(target, filename))
					Ω(exists).ShouldNot(BeTrue())
				})

				It("Should have a non nil error and not panic", func() {
					var err error
					Ω(func() {
						err = er.RunDbAction([]SystemDump{info["ConsoledbInfo"]}, ExportArchive)
					}).ShouldNot(Panic())
					Ω(err).ShouldNot(BeNil())
				})
			})

			Context("Restore", func() {
				It("Should have a non nil error and not panic", func() {
					var err error
					Ω(func() {
						err = er.RunDbAction([]SystemDump{info["ConsoledbInfo"]}, ImportArchive)
					}).ShouldNot(Panic())
					Ω(err).ShouldNot(BeNil())
				})
			})
		})
	})
}
