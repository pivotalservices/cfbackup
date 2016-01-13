package cfbackup

import (
	"fmt"
	"io"
	"os"

	"github.com/pivotalservices/gtils/command"
	"github.com/pivotalservices/gtils/persistence"
	"github.com/xchapter7x/goutil"
)

func init() {
	SetPGDumpUtilVersions()
}

//SetPGDumpUtilVersions -- set version paths for pgdump/pgrestore utils
func SetPGDumpUtilVersions() {
	switch os.Getenv(ER_VERSION_ENV_FLAG) {
	case ER_VERSION_16:
		persistence.PGDMP_DUMP_BIN = "/var/vcap/packages/postgres-9.4.2/bin/pg_dump"
		persistence.PGDMP_RESTORE_BIN = "/var/vcap/packages/postgres-9.4.2/bin/pg_restore"
	default:
		persistence.PGDMP_DUMP_BIN = "/var/vcap/packages/postgres/bin/pg_dump"
		persistence.PGDMP_RESTORE_BIN = "/var/vcap/packages/postgres/bin/pg_restore"
	}
}

const (
	//SD_PRODUCT --
	SD_PRODUCT string = "Product"
	//SD_COMPONENT --
	SD_COMPONENT string = "Component"
	//SD_IDENTITY --
	SD_IDENTITY string = "Identity"
	//SD_IP --
	SD_IP string = "Ip"
	//SD_USER --
	SD_USER string = "User"
	//SD_PASS --
	SD_PASS string = "Pass"
	//SD_VCAPUSER --
	SD_VCAPUSER string = "VcapUser"
	//SD_VCAPPASS --
	SD_VCAPPASS string = "VcapPass"
)

type (
	//PersistanceBackup - a struct representing a persistence backup
	PersistanceBackup interface {
		Dump(io.Writer) error
		Import(io.Reader) error
	}

	stringGetterSetter interface {
		Get(string) string
		Set(string, string)
	}
	//SystemDump - definition for a SystemDump interface
	SystemDump interface {
		stringGetterSetter
		Error() error
		GetPersistanceBackup() (dumper PersistanceBackup, err error)
	}
	//SystemInfo - a struct representing a base systemdump implementation
	SystemInfo struct {
		goutil.GetSet
		Product   string
		Component string
		Identity  string
		Ip        string
		User      string
		Pass      string
		VcapUser  string
		VcapPass  string
	}
	//PgInfo - a struct representing a pgres systemdump implementation
	PgInfo struct {
		SystemInfo
		Database string
	}
	//MysqlInfo - a struct representing a mysql systemdump implementation
	MysqlInfo struct {
		SystemInfo
		Database string
	}
	//NfsInfo - a struct representing a nfs systemdump implementation
	NfsInfo struct {
		SystemInfo
	}
)

//Get - a getter for a systeminfo object
func (s *SystemInfo) Get(name string) string {
	return s.GetSet.Get(s, name).(string)
}

//Set - a setter for a systeminfo object
func (s *SystemInfo) Set(name string, val string) {
	s.GetSet.Set(s, name, val)
}

//GetPersistanceBackup - the constructor for a new nfsinfo object
func (s *NfsInfo) GetPersistanceBackup() (dumper PersistanceBackup, err error) {
	return NewNFSBackup(s.Pass, s.Ip)
}

//GetPersistanceBackup - the constructor for a new mysqlinfo object
func (s *MysqlInfo) GetPersistanceBackup() (dumper PersistanceBackup, err error) {
	sshConfig := command.SshConfig{
		Username: s.VcapUser,
		Password: s.VcapPass,
		Host:     s.Ip,
		Port:     22,
	}
	return persistence.NewRemoteMysqlDump(s.User, s.Pass, sshConfig)
}

//GetPersistanceBackup - the constructor for a new pginfo object
func (s *PgInfo) GetPersistanceBackup() (dumper PersistanceBackup, err error) {
	sshConfig := command.SshConfig{
		Username: s.VcapUser,
		Password: s.VcapPass,
		Host:     s.Ip,
		Port:     22,
	}
	return persistence.NewPgRemoteDump(2544, s.Database, s.User, s.Pass, sshConfig)
}

//GetPersistanceBackup - the constructor for a systeminfo object
func (s *SystemInfo) GetPersistanceBackup() (dumper PersistanceBackup, err error) {
	panic("you have to extend SystemInfo and implement GetPersistanceBackup method on the child")
	return
}

//Error - method making systeminfo implement the error interface
func (s *SystemInfo) Error() (err error) {
	if s.Product == "" ||
		s.Component == "" ||
		s.Identity == "" ||
		s.Ip == "" ||
		s.User == "" ||
		s.Pass == "" ||
		s.VcapUser == "" ||
		s.VcapPass == "" {
		err = fmt.Errorf("invalid or incomplete system info object: %+v", s)
	}
	return
}
