package cfbackup

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/pivotalservices/gtils/command"
	"github.com/pivotalservices/gtils/osutils"
	"github.com/xchapter7x/lo"
)

const (
	NFSBackupTypeFull = "full"
	NFSBackupTypeLite = "lite"
	NFSBackupTypeNone = "skip"
)

//NewNFSBackup - constructor for an nfsbackup object
func NewNFSBackup(username, password, ip, sslKey, remoteArchivePath, backupType string) (nfs *NFSBackup, err error) {
	config := command.SshConfig{
		Username: username,
		Password: password,
		Host:     ip,
		Port:     22,
		SSLKey:   sslKey,
	}
	var remoteExecuter command.Executer

	if remoteExecuter, err = NfsNewRemoteExecuter(config); err == nil {
		nfs = &NFSBackup{
			Caller:     remoteExecuter,
			RemoteOps:  osutils.NewRemoteOperationsWithPath(config, remoteArchivePath),
			BackupType: backupType,
		}
	}
	return
}

//Dump - will dump the output of a executed command to the given writer
func (s *NFSBackup) Dump(dest io.Writer) (err error) {
	err = s.Caller.Execute(dest, s.getDumpCommand())
	return
}

//Import - will upload the contents of the given io.reader to the remote execution target and execute the restore command against the uploaded file.
func (s *NFSBackup) Import(lfile io.Reader) (err error) {
	lo.G.Debug("uploading file for backup")
	if err = s.RemoteOps.UploadFile(lfile); err == nil {
		lo.G.Debug("starting backup from %s", s.RemoteOps.Path())
		err = s.Caller.Execute(ioutil.Discard, s.getRestoreCommand())
	}
	if err == nil {
		lo.G.Debug("backup from %s completed", s.RemoteOps.Path())
	} else {
		lo.G.Debug("backup from %s completed with error %s", s.RemoteOps.Path(), err)
	}
	s.RemoteOps.RemoveRemoteFile()
	return
}

func (s *NFSBackup) getRestoreCommand() string {
	return fmt.Sprintf("cd %s && tar zxf %s", NfsDirPath, s.RemoteOps.Path())
}

func (s *NFSBackup) getDumpCommand() string {
	var cmd string
	switch s.BackupType {
	case NFSBackupTypeLite:
		cmd = fmt.Sprintf("cd %s && tar cz --exclude=cc-resources %s", NfsDirPath, NfsArchiveDir)
	case NFSBackupTypeNone:
		cmd = fmt.Sprintf("cd %s && tar cz --include=*/cc-buildpacks/* %s/*", NfsDirPath, NfsArchiveDir)
	case NFSBackupTypeFull:
		cmd = fmt.Sprintf("cd %s && tar cz %s", NfsDirPath, NfsArchiveDir)
	}

	return cmd
}
