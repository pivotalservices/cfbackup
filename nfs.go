package cfbackup

import (
	"io"

	"github.com/pivotalservices/gtils/command"
	"github.com/pivotalservices/gtils/osutils"
)

type remoteOperationsInterface interface {
	UploadFile(lfile io.Reader) (err error)
	Path() string
}

func BackupNfs(password, ip string, dest io.Writer) (err error) {
	var nfsb *NFSBackup

	if nfsb, err = NewNFSBackup(password, ip); err == nil {
		err = nfsb.Dump(dest)
	}
	return
}

type NFSBackup struct {
	Caller    command.Executer
	RemoteOps remoteOperationsInterface
}

var NfsNewRemoteExecuter func(command.SshConfig) (command.Executer, error) = command.NewRemoteExecutor

func NewNFSBackup(password, ip string) (nfs *NFSBackup, err error) {
	config := command.SshConfig{
		Username: "vcap",
		Password: password,
		Host:     ip,
		Port:     22,
	}
	var remoteExecuter command.Executer

	if remoteExecuter, err = NfsNewRemoteExecuter(config); err == nil {
		nfs = &NFSBackup{
			Caller:    remoteExecuter,
			RemoteOps: osutils.NewRemoteOperations(config),
		}
	}
	return
}

func (s *NFSBackup) Import(lfile io.Reader) (err error) {
	if err = s.RemoteOps.UploadFile(lfile); err == nil {
		//err = s.restore()
	}
	return
}

func (s *NFSBackup) getRestoreCommand() string {
	return "cd /var/vcap/store && tar zx file.tgz"
}

func (s *NFSBackup) Dump(dest io.Writer) (err error) {
	command := "cd /var/vcap/store && tar cz shared"
	err = s.Caller.Execute(dest, command)
	return
}
