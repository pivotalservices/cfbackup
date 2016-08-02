package elasticruntime

import (
	"os"

	"github.com/pivotalservices/cfbackup"
	ghttp "github.com/pivotalservices/gtils/http"
)

type (

	//ElasticRuntime contains information about a Pivotal Elastic Runtime deployment
	ElasticRuntime struct {
		cfbackup.BackupContext
		JSONFile          string
		SystemsInfo       cfbackup.SystemsInfo
		PersistentSystems []cfbackup.SystemDump
		HTTPGateway       ghttp.HttpGateway
		InstallationName  string
		SSHPrivateKey     string
		NFS               string
	}

	//ElasticRuntimeBuilder -- an object that can build an elastic runtime pre-initialized
	ElasticRuntimeBuilder struct{}

	//TempFile -- a wrapper around temp files to make a closer
	TempFile struct {
		FileRef *os.File
	}
)
