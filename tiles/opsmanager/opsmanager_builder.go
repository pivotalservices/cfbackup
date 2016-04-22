package opsmanager

import (
	"github.com/pivotalservices/cfbackup"
	"github.com/pivotalservices/cfbackup/tileregistry"
	"github.com/pivotalservices/gtils/command"
	"github.com/pivotalservices/opsmanclient"
	"github.com/xchapter7x/lo"
)

//New -- builds a new ops manager object pre initialized
func (s *OpsManagerBuilder) New(tileSpec tileregistry.TileSpec) (opsManagerTileCloser tileregistry.TileCloser, err error) {
	var opsManager *OpsManager
	opsManager, err = NewOpsManager(tileSpec.OpsManagerHost, tileSpec.AdminUser, tileSpec.AdminPass, tileSpec.OpsManagerUser, tileSpec.OpsManagerPass, tileSpec.OpsManagerPassphrase, tileSpec.ArchiveDirectory, tileSpec.CryptKey)
	opsManager.ClearBoshManifest = tileSpec.ClearBoshManifest

	if installationSettings, err := opsManager.GetInstallationSettings(); err == nil {
		config := cfbackup.NewConfigurationParserFromReader(installationSettings)

		if iaas, hasKey := config.GetIaaS(); hasKey {
			lo.G.Debug("we found a iaas info block")
			var executor command.Executer
			if executor, err = opsmanclient.NewSSHExecuter(tileSpec.OpsManagerUser, tileSpec.OpsManagerPass, tileSpec.OpsManagerHost, iaas.SSHPrivateKey, 22); err == nil {
				opsManager.setSSHExecutor(executor)
			}

		} else {
			lo.G.Debug("No IaaS PEM key found. Defaulting to using ssh username and password credentials")
		}
	}
	opsManagerTileCloser = struct {
		tileregistry.Tile
		tileregistry.Closer
	}{
		opsManager,
		new(tileregistry.DoNothingCloser),
	}
	return
}
