package opsmanager

import (
	"io"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/pivotalservices/cfbackup"
	"github.com/pivotalservices/gtils/command"
	"github.com/pivotalservices/opsmanclient"
	"github.com/xchapter7x/lo"
)

// NewOpsManager initializes an OpsManager instance
var NewOpsManager = func(opsManagerHostname string, adminUsername string, adminPassword string, opsManagerUsername string, opsManagerPassword string, opsManagerPassphrase string, target string, cryptKey string) (context *OpsManager, err error) {
	backupContext := cfbackup.NewBackupContext(target, cfenv.CurrentEnv(), cryptKey)
	opsmanClient := opsmanclient.New("https://"+opsManagerHostname, adminUsername, adminPassword, opsManagerPassphrase, backupContext.IsS3)

	context = &OpsManager{
		BackupContext: backupContext,
		Client:        opsmanClient,
	}
	context.Executor, err = opsmanclient.NewSSHExecuter(opsManagerUsername, opsManagerPassword, opsManagerHostname, "", 22)
	return
}

// SetSSHExecutor - sets the remote ssh executer associated with the opsmanager
func (context *OpsManager) setSSHExecutor(executor command.Executer) {
	lo.G.Debug("Setting SSH Executor")
	context.Executor = executor
}

// GetInstallationSettings retrieves all the installation settings from OpsMan
// and returns them in a buffered reader
func (context *OpsManager) GetInstallationSettings() (settings io.Reader, err error) {
	return context.Client.GetInstallationSettingsBuffered()
}

//~ Backup Operations

// Backup performs a backup of a Pivotal Ops Manager instance
func (context *OpsManager) Backup() (err error) {
	if backupWriter, err := context.Writer(context.BackupContext.TargetDir, context.OpsmanagerBackupDir, OpsMgrDeploymentsFileName); err == nil {
		if err = context.saveDeployments(backupWriter); err == nil {
			return context.saveInstallation(backupWriter)
		}
	}
	return
}

func (context *OpsManager) saveDeployments(backupWriter io.WriteCloser) (err error) {
	return context.Client.SaveDeployments(context.Executor, backupWriter)
}

func (context *OpsManager) saveInstallation(backupWriter io.WriteCloser) error {
	return context.Client.SaveInstallation(backupWriter)
}

//~ Restore Operations

// Restore performs a restore of a Pivotal Ops Manager instance
func (context *OpsManager) Restore() error {
	if backupReader, err := context.Reader(context.TargetDir, context.OpsmanagerBackupDir, OpsMgrInstallationAssetsFileName); err == nil {
		return context.Client.ImportInstallation(context.Executor, context.OpsmanagerBackupDir, backupReader, context.ClearBoshManifest)
	}
	return nil
}
