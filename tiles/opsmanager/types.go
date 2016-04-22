package opsmanager

import (
	"io"
	"net/http"

	"github.com/pivotalservices/cfbackup"
	"github.com/pivotalservices/gtils/command"
	ghttp "github.com/pivotalservices/gtils/http"
)

type (

	// OpsManager contains the location and credentials of a Pivotal Ops Manager instance
	OpsManager struct {
		cfbackup.BackupContext
		Executor command.Executer
		Client
		// Hostname            string
		// Username            string
		// Password            string
		// Passphrase          string
		// TempestPassword     string
		// DbEncryptionKey     string
		// LocalExecuter       command.Executer
		// SettingsUploader    httpUploader
		// AssetsUploader      httpUploader
		// SettingsRequestor   httpRequestor
		// AssetsRequestor     httpRequestor
		// DeploymentDir       string
		OpsmanagerBackupDir string
		// SSHPrivateKey       string
		// SSHUsername         string
		// SSHPassword         string
		// SSHPort             int
		ClearBoshManifest bool
	}

	// Client used to call the Ops Manager API
	Client interface {
		GetInstallationSettingsBuffered() (io.Reader, error)
		SaveDeployments(e command.Executer, bw io.WriteCloser) error
		SaveInstallation(backupWriter io.WriteCloser) error
		ImportInstallation(e command.Executer, backupDir string, backupReader io.ReadCloser, removeBoshManifest bool) error
	}

	//OpsManagerBuilder - an object that can build ops manager objects
	OpsManagerBuilder struct{}

	httpUploader func(conn ghttp.ConnAuth, paramName, filename string, fileSize int64, fileRef io.Reader, params map[string]string) (res *http.Response, err error)

	httpRequestor interface {
		Get(ghttp.HttpRequestEntity) ghttp.RequestAdaptor
		Post(ghttp.HttpRequestEntity, io.Reader) ghttp.RequestAdaptor
		Put(ghttp.HttpRequestEntity, io.Reader) ghttp.RequestAdaptor
	}
)
