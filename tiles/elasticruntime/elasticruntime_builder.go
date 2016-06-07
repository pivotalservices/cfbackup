package elasticruntime

import (
	"io"

	"github.com/pivotalservices/cfbackup"
	"github.com/pivotalservices/cfbackup/tileregistry"
	"github.com/pivotalservices/cfbackup/tiles/opsmanager"
)

//New -- method to generate an initialized elastic runtime
func (s *ElasticRuntimeBuilder) New(tileSpec tileregistry.TileSpec) (elasticRuntimeCloser tileregistry.TileCloser, err error) {
	var (
		installationSettings io.Reader
		tmpfile              *TempFile
		sshKey               = ""
	)
	if tmpfile, err = NewTempFile(opsmanager.OpsMgrInstallationSettingsFilename); err == nil {

		if installationSettings, err = GetInstallationSettings(tileSpec); err == nil {
			io.Copy(tmpfile.FileRef, installationSettings)
			config := cfbackup.NewConfigurationParser(tmpfile.FileRef.Name())

			if iaas, hasKey := config.GetIaaS(); hasKey {
				sshKey = iaas.SSHPrivateKey
			}
			elasticRuntime := NewElasticRuntime(tmpfile.FileRef.Name(), tileSpec.ArchiveDirectory, sshKey, tileSpec.CryptKey)
			elasticRuntimeCloser = struct {
				tileregistry.Tile
				tileregistry.Closer
			}{
				elasticRuntime,
				tmpfile,
			}
		}
	}
	return
}

//GetInstallationSettings - makes a call to ops manager and returns a io.reader containing the contents of the installation settings file.
var GetInstallationSettings = func(tileSpec tileregistry.TileSpec) (settings io.Reader, err error) {
	var (
		opsManager *opsmanager.OpsManager
	)

	if opsManager, err = opsmanager.NewOpsManager(tileSpec.OpsManagerHost, tileSpec.AdminUser, tileSpec.AdminPass, tileSpec.OpsManagerUser, tileSpec.OpsManagerPass, tileSpec.OpsManagerPassphrase, tileSpec.ArchiveDirectory, tileSpec.CryptKey); err == nil {
		settings, err = opsManager.GetInstallationSettings()
	}
	return
}
