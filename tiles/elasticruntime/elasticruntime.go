package elasticruntime

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/pivotalservices/cfbackup"
	"github.com/xchapter7x/lo"

	errwrap "github.com/pkg/errors"
)

// NewElasticRuntime initializes an ElasticRuntime intance
var NewElasticRuntime = func(jsonFile string, target string, sshKey string, cryptKey string, nfs string) *ElasticRuntime {

	if _, err := os.Stat(jsonFile); err != nil {
		lo.G.Error("installation settings not found: ", err)
		lo.G.Panic("exiting program, cant work without a valid installation settings...")
	}
	systemsInfo := cfbackup.NewSystemsInfo(jsonFile, sshKey, nfs)
	context := &ElasticRuntime{
		SSHPrivateKey:     sshKey,
		JSONFile:          jsonFile,
		BackupContext:     cfbackup.NewBackupContext(target, cfenv.CurrentEnv(), cryptKey),
		SystemsInfo:       systemsInfo,
		PersistentSystems: systemsInfo.PersistentSystems(),
	}
	return context
}

// Backup performs a backup of a Pivotal Elastic Runtime deployment
func (context *ElasticRuntime) Backup() (err error) {
	return context.backupRestore(cfbackup.ExportArchive)
}

// Restore performs a restore of a Pivotal Elastic Runtime deployment
func (context *ElasticRuntime) Restore() (err error) {
	err = context.backupRestore(cfbackup.ImportArchive)
	return
}

func (context *ElasticRuntime) backupRestore(action int) (err error) {
	var (
		ccJobs []cfbackup.CCJob
	)

	err = context.ReadAllUserCredentials()
	if err != nil {
		return errwrap.Wrap(err, "failed reading user credentials")
	}

	directorCredentialsValid, err := context.directorCredentialsValid()
	if err != nil {
		return errwrap.Wrap(err, "failed on check for valid director credentials")
	}

	if directorCredentialsValid {
		lo.G.Debug("Retrieving All CC VMs")
		manifest, erro := context.getManifest()
		if err != nil {
			return erro
		}
		if ccJobs, err = context.getAllCloudControllerVMs(); err == nil {
			directorInfo := context.SystemsInfo.SystemDumps[cfbackup.ERDirector]
			var cloudController *cfbackup.CloudController
			cloudController, err = cfbackup.NewCloudController(directorInfo.Get(cfbackup.SDIP), directorInfo.Get(cfbackup.SDUser), directorInfo.Get(cfbackup.SDPass), context.InstallationName, manifest, ccJobs)
			if err != nil {
				return errwrap.Wrap(err, "failed creating new cloud controller")
			}

			lo.G.Debug("Setting up CC jobs")
			defer cloudController.Start()
			if err := cloudController.Stop(); err != nil {
				return errwrap.Wrap(err, "failed to stop cloud controller")
			}
		} else {
			return errwrap.Wrap(err, "failed getting VMs")
		}

		lo.G.Debug("Running db action")
		if len(context.PersistentSystems) > 0 {
			err = context.RunDbAction(context.PersistentSystems, action)
			if err != nil {
				lo.G.Error("Error backing up db", err)
				err = ErrERDBBackup
			}
		} else {
			lo.G.Info("There is no internal persistent system used by ERT, skip db action")
		}
	} else if err == nil {
		err = cfbackup.ErrERDirectorCreds
	}
	return
}

func (context *ElasticRuntime) getAllCloudControllerVMs() (ccvms []cfbackup.CCJob, err error) {

	var jsonObj []cfbackup.VMObject
	lo.G.Debug("Entering getAllCloudControllerVMs() function")
	directorInfo := context.SystemsInfo.SystemDumps[cfbackup.ERDirector]

	director, err := cfbackup.NewDirector(directorInfo.Get(cfbackup.SDIP), directorInfo.Get(cfbackup.SDUser), directorInfo.Get(cfbackup.SDPass), 25555)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed creating new director")
	}

	body, err := director.GetCloudControllerVMSet(context.InstallationName)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed to get cloud controller vm set")
	}
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed to read response body")
	}

	err = json.Unmarshal(bodyBytes, &jsonObj)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed to unmarshal response body")
	}

	if len(jsonObj) <= 0 {
		return nil, fmt.Errorf("no vm objects found: %v bytes: %v", jsonObj, string(bodyBytes))
	}
	return cfbackup.GetCCVMs(jsonObj)
}

//RunDbAction - run a db action dump/import against a list of systemdump types
func (context *ElasticRuntime) RunDbAction(dbInfoList []cfbackup.SystemDump, action int) (err error) {

	for _, info := range dbInfoList {
		lo.G.Debug(fmt.Sprintf("RunDbAction info: %v %v", info.Get("Product"), info.Get("Component")))

		if err = info.Error(); err == nil {
			err = context.readWriterArchive(info, context.TargetDir, action)

		} else {
			lo.G.Error("readWriterArchive err: ", err)
			break
		}
	}
	return
}

func (context *ElasticRuntime) readWriterArchive(dbInfo cfbackup.SystemDump, databaseDir string, action int) (err error) {
	filename := fmt.Sprintf(ERBackupFileFormat, dbInfo.Get(cfbackup.SDComponent))
	filepath := path.Join(databaseDir, filename)

	var pb cfbackup.PersistanceBackup

	if pb, err = dbInfo.GetPersistanceBackup(); err == nil {
		switch action {
		case cfbackup.ImportArchive:
			lo.G.Debug("Restoring %s", dbInfo.Get(cfbackup.SDComponent))
			var backupReader io.ReadCloser
			if backupReader, err = context.Reader(filepath); err == nil {
				defer backupReader.Close()
				err = pb.Import(backupReader)
				lo.G.Debug("Done restoring %s", dbInfo.Get(cfbackup.SDComponent))
			}
		case cfbackup.ExportArchive:
			lo.G.Info("Exporting %s", dbInfo.Get(cfbackup.SDComponent))
			var backupWriter io.WriteCloser
			if backupWriter, err = context.Writer(filepath); err == nil {
				defer backupWriter.Close()
				err = pb.Dump(backupWriter)
				lo.G.Debug("Done backing up ", dbInfo.Get(cfbackup.SDComponent), err)
			}
		}
	}
	return
}

//ReadAllUserCredentials - get all user creds from the installation json
func (context *ElasticRuntime) ReadAllUserCredentials() (err error) {
	configParser := cfbackup.NewConfigurationParser(context.JSONFile)
	installationSettings := configParser.InstallationSettings
	err = context.assignCredentialsAndInstallationName(installationSettings)
	return
}

func (context *ElasticRuntime) assignCredentialsAndInstallationName(installationSettings cfbackup.InstallationSettings) (err error) {

	if err = context.assignCredentials(installationSettings); err == nil {
		context.InstallationName, err = context.getDeploymentName(installationSettings)
	}
	return
}

func (context *ElasticRuntime) assignCredentials(installationSettings cfbackup.InstallationSettings) (err error) {

	for name, sysInfo := range context.SystemsInfo.SystemDumps {
		var (
			userID string
			ip     string
			pass   string
		)
		productName := sysInfo.Get(cfbackup.SDProduct)
		jobName := sysInfo.Get(cfbackup.SDComponent)
		identifier := sysInfo.Get(cfbackup.SDIdentifier)

		if userID, pass, ip, err = context.getVMUserIDPasswordAndIP(installationSettings, productName, jobName); err == nil {
			sysInfo.Set(cfbackup.SDIP, ip)
			sysInfo.Set(cfbackup.SDVcapPass, pass)
			sysInfo.Set(cfbackup.SDVcapUser, userID)
			if identifier == "vm_credentials" {
				sysInfo.Set(cfbackup.SDUser, userID)
				sysInfo.Set(cfbackup.SDPass, pass)
			} else if userID, pass, err = context.getUserIDPasswordForIdentifier(installationSettings, productName, jobName, identifier); err == nil {
				sysInfo.Set(cfbackup.SDUser, userID)
				sysInfo.Set(cfbackup.SDPass, pass)
			}
		}
		context.SystemsInfo.SystemDumps[name] = sysInfo
	}
	return
}

func (context *ElasticRuntime) directorCredentialsValid() (ok bool, err error) {
	var directorInfo cfbackup.SystemDump

	directorInfo, ok = context.SystemsInfo.SystemDumps[cfbackup.ERDirector]
	if !ok {
		return false, ErrERDirectorCreds
	}

	director, err := cfbackup.NewDirector(directorInfo.Get(cfbackup.SDIP), directorInfo.Get(cfbackup.SDUser), directorInfo.Get(cfbackup.SDPass), 25555)
	if err != nil {
		return false, errwrap.Wrap(err, "failed creating new director")
	}

	_, err = director.GetInfo()
	if err != nil {
		return false, errwrap.Wrap(err, "failed to get info from director")
	}
	return true, nil
}

func (context *ElasticRuntime) getManifest() (manifest []byte, err error) {
	directorInfo, _ := context.SystemsInfo.SystemDumps[cfbackup.ERDirector]
	director, err := cfbackup.NewDirector(directorInfo.Get(cfbackup.SDIP), directorInfo.Get(cfbackup.SDUser), directorInfo.Get(cfbackup.SDPass), 25555)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed creating new director")
	}
	mfs, err := director.GetDeploymentManifest(context.InstallationName)
	if err != nil {
		return nil, errwrap.Wrap(err, fmt.Sprintf("failed on GetDeploymentManifest for %s", context.InstallationName))
	}
	data, err := ioutil.ReadAll(mfs)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed reading response body")
	}
	return data, nil
}

func (context *ElasticRuntime) getDeploymentName(installationSettings cfbackup.InstallationSettings) (deploymentName string, err error) {
	var product cfbackup.Products
	if product, err = installationSettings.FindByProductID("cf"); err == nil {
		deploymentName = product.InstallationName
	}
	return
}

func (context *ElasticRuntime) getUserIDPasswordForIdentifier(installationSettings cfbackup.InstallationSettings, product, component, identifier string) (userID, password string, err error) {
	var propertyMap map[string]string
	if propertyMap, err = installationSettings.FindPropertyValues(product, component, identifier); err == nil {
		userID = propertyMap["identity"]
		password = propertyMap["password"]
	}
	return
}

func (context *ElasticRuntime) getVMUserIDPasswordAndIP(installationSettings cfbackup.InstallationSettings, product, component string) (userID, password, ip string, err error) {
	var ips []string
	if ips, err = installationSettings.FindIPsByProductAndJob(product, component); err == nil {
		if len(ips) > 0 {
			ip = ips[0]
		} else {
			err = fmt.Errorf("No IPs found for %s, %s", product, component)
		}
		var vmCredential cfbackup.VMCredentials
		if vmCredential, err = installationSettings.FindVMCredentialsByProductAndJob(product, component); err == nil {
			userID = vmCredential.UserID
			password = vmCredential.Password
		}
	}
	return
}
