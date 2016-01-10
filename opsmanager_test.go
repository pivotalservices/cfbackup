package cfbackup_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/cloudfoundry-community/go-cfenv"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/cfbackup"
	. "github.com/pivotalservices/gtils/command"
	ghttp "github.com/pivotalservices/gtils/http"
	"github.com/pivotalservices/gtils/http/fake"
	"github.com/pivotalservices/gtils/osutils"
)

var _ = Describe("OpsManager object", func() {
	var (
		opsManager *OpsManager
		tmpDir     string
		backupDir  string
	)
	Describe("Given a GetInstallationSettings method", func() {
		Context("when called on a properly initialized opsmanager", func() {
			var (
				installationSettings        io.Reader
				err                         error
				controlInstallationSettings = "{my:fake, installation:settings, object:blob}"
			)
			BeforeEach(func() {
				tmpDir, _ = ioutil.TempDir("/tmp", "test")
				backupDir = path.Join(tmpDir, "backup", "opsmanager")
				gw := &MockHttpGateway{StatusCode: 200, State: controlInstallationSettings}

				opsManager = &OpsManager{
					SettingsUploader:    MockMultiPartUploadFunc,
					AssetsUploader:      MockMultiPartUploadFunc,
					SettingsRequestor:   gw,
					AssetsRequestor:     gw,
					Hostname:            "localhost",
					Username:            "user",
					Password:            "password",
					BackupContext:       NewFakeBackupContext(path.Join(tmpDir, "backup"), cfenv.CurrentEnv(), new(DiskProvider)),
					Executer:            &successExecuter{},
					DeploymentDir:       "fixtures/encryptionkey",
					OpsmanagerBackupDir: "opsmanager",
				}
				installationSettings, err = opsManager.GetInstallationSettings()
			})
			It("then it should return the installation_settings json from the ops manager api", func() {
				contents, _ := ioutil.ReadAll(installationSettings)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(contents).Should(Equal([]byte(controlInstallationSettings)))
			})
		})
	})
	Describe("Given a Restore method", func() {

		Context("when called", func() {
			var (
				opsMgr                  *OpsManager
				fakeSettingsUploader    *fake.MultiPart
				fakeAssetsUploader      *fake.MultiPart
				controlAssetsContents   = []byte(`test assets`)
				controlSettingsContents = []byte(`test bytes`)
			)

			BeforeEach(func() {
				fakeSettingsUploader = new(fake.MultiPart)
				fakeAssetsUploader = new(fake.MultiPart)
				tmpDir, _ = ioutil.TempDir("/tmp", "test")
				backupDir = path.Join(tmpDir, "backup", "opsmanager")
				gw := &MockHttpGateway{}

				opsMgr = &OpsManager{
					SettingsUploader:    fakeSettingsUploader.Upload,
					AssetsUploader:      fakeAssetsUploader.Upload,
					SettingsRequestor:   gw,
					AssetsRequestor:     gw,
					Hostname:            "localhost",
					Username:            "user",
					Password:            "password",
					BackupContext:       NewFakeBackupContext(path.Join(tmpDir, "backup"), cfenv.CurrentEnv(), new(DiskProvider)),
					Executer:            &successExecuter{},
					DeploymentDir:       "fixtures/encryptionkey",
					OpsmanagerBackupDir: "opsmanager",
				}
				f, _ := osutils.SafeCreate(opsMgr.TargetDir, opsMgr.OpsmanagerBackupDir, OpsMgrInstallationSettingsFilename)
				f.Write(controlSettingsContents)
				f.Close()
				f, _ = osutils.SafeCreate(opsMgr.TargetDir, opsMgr.OpsmanagerBackupDir, OpsMgrInstallationAssetsFileName)
				f.Write(controlAssetsContents)
				f.Close()
				opsMgr.Restore()
			})
			It("then it should import the assets archive", func() {
				Ω(fakeAssetsUploader.UploadCallCount).ShouldNot(Equal(0))
				Ω(fakeAssetsUploader.SpyFileContents).Should(Equal(controlAssetsContents))
			})
			It("then it should not import the settings archive", func() {
				Ω(fakeSettingsUploader.UploadCallCount).Should(Equal(0))
				Ω(fakeSettingsUploader.SpyFileContents).Should(BeNil())
			})
		})

		Context("calling restore with failed removal of deployment files", func() {

			BeforeEach(func() {
				tmpDir, _ = ioutil.TempDir("/tmp", "test")
				backupDir = path.Join(tmpDir, "backup", "opsmanager")
				gw := &MockHttpGateway{}

				opsManager = &OpsManager{
					SettingsUploader:    MockMultiPartUploadFunc,
					AssetsUploader:      MockMultiPartUploadFunc,
					SettingsRequestor:   gw,
					AssetsRequestor:     gw,
					Hostname:            "localhost",
					Username:            "user",
					Password:            "password",
					BackupContext:       NewFakeBackupContext(path.Join(tmpDir, "backup"), cfenv.CurrentEnv(), new(DiskProvider)),
					Executer:            &failExecuter{},
					DeploymentDir:       "fixtures/encryptionkey",
					OpsmanagerBackupDir: "opsmanager",
				}
				f, _ := osutils.SafeCreate(opsManager.TargetDir, opsManager.OpsmanagerBackupDir, OpsMgrInstallationSettingsFilename)
				f.Close()
				f, _ = osutils.SafeCreate(opsManager.TargetDir, opsManager.OpsmanagerBackupDir, OpsMgrInstallationAssetsFileName)
				f.Close()
			})

			It("Should yield error", func() {
				err := opsManager.Restore()
				Ω(err).ShouldNot(BeNil())
			})
		})

		Context("calling restore successfully", func() {

			BeforeEach(func() {
				tmpDir, _ = ioutil.TempDir("/tmp", "test")
				backupDir = path.Join(tmpDir, "backup", "opsmanager")
				gw := &MockHttpGateway{}

				opsManager = &OpsManager{
					SettingsUploader:    MockMultiPartUploadFunc,
					AssetsUploader:      MockMultiPartUploadFunc,
					SettingsRequestor:   gw,
					AssetsRequestor:     gw,
					Hostname:            "localhost",
					Username:            "user",
					Password:            "password",
					BackupContext:       NewFakeBackupContext(path.Join(tmpDir, "backup"), cfenv.CurrentEnv(), new(DiskProvider)),
					Executer:            &successExecuter{},
					DeploymentDir:       "fixtures/encryptionkey",
					OpsmanagerBackupDir: "opsmanager",
				}
				f, _ := osutils.SafeCreate(opsManager.TargetDir, opsManager.OpsmanagerBackupDir, OpsMgrInstallationSettingsFilename)
				f.Close()
				f, _ = osutils.SafeCreate(opsManager.TargetDir, opsManager.OpsmanagerBackupDir, OpsMgrInstallationAssetsFileName)
				f.Close()
			})

			It("Should yield nil error", func() {
				err := opsManager.Restore()
				Ω(err).Should(BeNil())
			})
		})

		Context("calling restore when unable to upload", func() {
			BeforeEach(func() {
				tmpDir, _ = ioutil.TempDir("/tmp", "test")
				backupDir = path.Join(tmpDir, "backup", "opsmanager")
				gw := &MockHttpGateway{StatusCode: 500, State: failureString}

				opsManager = &OpsManager{
					SettingsUploader:    ghttp.MultiPartUpload,
					AssetsUploader:      ghttp.MultiPartUpload,
					SettingsRequestor:   gw,
					AssetsRequestor:     gw,
					Hostname:            "localhost",
					Username:            "user",
					Password:            "password",
					BackupContext:       NewFakeBackupContext(path.Join(tmpDir, "backup"), cfenv.CurrentEnv(), new(DiskProvider)),
					Executer:            &failExecuter{},
					DeploymentDir:       "fixtures/encryptionkey",
					OpsmanagerBackupDir: "opsmanager",
				}
				f, _ := osutils.SafeCreate(opsManager.TargetDir, opsManager.OpsmanagerBackupDir, OpsMgrInstallationSettingsFilename)
				f.Close()
			})

			It("Should yield a non-nil error", func() {
				err := opsManager.Restore()
				Ω(err).ShouldNot(BeNil())
			})
		})

		Context("calling restore unsuccessfully", func() {
			BeforeEach(func() {
				tmpDir, _ = ioutil.TempDir("/tmp", "test")
				backupDir = path.Join(tmpDir, "backup", "opsmanager")
				gw := &MockHttpGateway{StatusCode: 500, State: failureString}

				opsManager = &OpsManager{
					SettingsUploader:    MockMultiPartUploadFunc,
					AssetsUploader:      MockMultiPartUploadFunc,
					SettingsRequestor:   gw,
					AssetsRequestor:     gw,
					Hostname:            "localhost",
					Username:            "user",
					Password:            "password",
					BackupContext:       NewFakeBackupContext(path.Join(tmpDir, "backup"), cfenv.CurrentEnv(), new(DiskProvider)),
					Executer:            &failExecuter{},
					DeploymentDir:       "fixtures/encryptionkey",
					OpsmanagerBackupDir: "opsmanager",
				}
				f, _ := osutils.SafeCreate(opsManager.TargetDir, opsManager.OpsmanagerBackupDir, OpsMgrInstallationSettingsFilename)
				f.Close()
			})

			It("Should yield a non-nil error", func() {
				err := opsManager.Restore()
				Ω(err).ShouldNot(BeNil())
			})
		})
	})

	Describe("Given a Backup method", func() {
		Context("When calling an ops manager api endpoint that returns a non successful status code", func() {
			var (
				err         error
				controlBody = "there was an error"
			)

			BeforeEach(func() {
				tmpDir, _ = ioutil.TempDir("/tmp", "test")
				backupDir = path.Join(tmpDir, "backup", "opsmanager")
				fakeHttpRequestor := &MockHttpGateway{StatusCode: http.StatusUnauthorized, State: controlBody}

				opsManager = &OpsManager{
					SettingsRequestor:   fakeHttpRequestor,
					Hostname:            "localhost",
					Username:            "user",
					Password:            "password",
					BackupContext:       NewFakeBackupContext(path.Join(tmpDir, "backup"), cfenv.CurrentEnv(), new(DiskProvider)),
					Executer:            &successExecuter{},
					LocalExecuter:       NewLocalMockExecuter(),
					DeploymentDir:       "fixtures/encryptionkey",
					OpsmanagerBackupDir: "opsmanager",
				}
				err = opsManager.Backup()
			})

			It("Then we should fail and output the error message from ops manager", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err.Error()).Should(Equal(controlBody))
			})
		})

		Context("called yielding an error in the chain", func() {
			BeforeEach(func() {
				tmpDir, _ = ioutil.TempDir("/tmp", "test")
				backupDir = path.Join(tmpDir, "backup", "opsmanager")

				opsManager = &OpsManager{
					Hostname:            "localhost",
					Username:            "user",
					Password:            "password",
					BackupContext:       NewFakeBackupContext(path.Join(tmpDir, "backup"), cfenv.CurrentEnv(), new(DiskProvider)),
					Executer:            &failExecuter{},
					LocalExecuter:       NewLocalMockExecuter(),
					DeploymentDir:       "fixtures/encryptionkey",
					OpsmanagerBackupDir: "opsmanager",
				}

			})

			It("should return non nil error and not write installation.json", func() {
				err := opsManager.Backup()
				filepath := path.Join(backupDir, "installation.json")
				Ω(err).ShouldNot(BeNil())
				Ω(osutils.Exists(filepath)).Should(BeFalse())
			})

			It("should return non nil error and not write deployments.tar.gz", func() {
				err := opsManager.Backup()
				filepath := path.Join(backupDir, "deployments.tar.gz")
				Ω(err).ShouldNot(BeNil())
				Ω(osutils.Exists(filepath)).Should(BeTrue())
			})
		})

		Context("called yielding a successful rest call", func() {

			BeforeEach(func() {
				tmpDir, _ = ioutil.TempDir("/tmp", "test")
				backupDir = path.Join(tmpDir, "backup", "opsmanager")
				gw := &MockHttpGateway{StatusCode: 200, State: successString}
				opsManager = &OpsManager{
					SettingsRequestor:   gw,
					Hostname:            "localhost",
					Username:            "user",
					Password:            "password",
					BackupContext:       NewFakeBackupContext(path.Join(tmpDir, "backup"), cfenv.CurrentEnv(), new(DiskProvider)),
					Executer:            &successExecuter{},
					LocalExecuter:       NewLocalMockExecuter(),
					DeploymentDir:       "fixtures/encryptionkey",
					OpsmanagerBackupDir: "opsmanager",
				}
			})

			It("should return nil error and write the proper information to the installation.json", func() {
				err := opsManager.Backup()
				filepath := path.Join(backupDir, "installation.json")
				b, _ := ioutil.ReadFile(filepath)
				Ω(err).Should(BeNil())
				Ω(b).Should(Equal([]byte(successString)))
			})

			It("should return nil error and write ", func() {
				opsManager.Backup()
				filepath := path.Join(backupDir, "deployments.tar.gz")
				Ω(osutils.Exists(filepath)).Should(BeTrue())
			})
		})
	})
})
