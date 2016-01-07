package cfbackup

import (
	"io"
	"os"
	"strings"

	"github.com/pivotalservices/gtils/log"
	"github.com/pivotalservices/gtils/osutils"
	"github.com/pivotalservices/gtils/storage"
	"github.com/xchapter7x/lo"
)

func NewBackupContext(targetDir string, env map[string]string) (backupContext BackupContext) {
	backupContext = BackupContext{
		TargetDir: targetDir,
	}

	if useS3(env) {
		backupContext.AccessKeyID = env[AccessKeyIDVarname]
		backupContext.SecretAccessKey = env[SecretAccessKeyVarname]
		backupContext.BucketName = env[BucketNameVarname]
		backupContext.S3Domain = env[S3Domain]
		backupContext.IsS3Archive = true
	}
	return
}

func useS3(env map[string]string) bool {
	_, akid := env[AccessKeyIDVarname]
	_, sak := env[SecretAccessKeyVarname]
	_, bn := env[BucketNameVarname]
	s3val, is := env[IsS3Varname]
	isS3 := (s3val == "true")
	return (akid && sak && bn && is && isS3)
}

func (s *BackupContext) Writer(path ...string) (io.WriteCloser, error) {
	if s.IsS3Archive {
		s3FileName := strings.Join(path, "/")
		lo.G.Debug("BackupContext.Writer()", log.Data{"s3FileName": s3FileName})

		s3, err := storage.SafeCreateS3Bucket(s.S3Domain, s.BucketName, s.AccessKeyID, s.SecretAccessKey)
		if err != nil {
			return nil, err
		}
		lo.G.Debug("Created Bucket", log.Data{"s3Domain": s.S3Domain, "bucketName": s.BucketName, "accessKeyID": s.AccessKeyID, "secretAccessKey": s.SecretAccessKey})
		return s3.NewWriter(s3FileName)
	}
	return osutils.SafeCreate(path...)
}

func (s *BackupContext) Reader(path ...string) (io.ReadCloser, error) {
	if s.IsS3Archive {
		s3, err := storage.SafeCreateS3Bucket(s.S3Domain, s.BucketName, s.AccessKeyID, s.SecretAccessKey)
		if err != nil {
			return nil, err
		}
		return s3.NewReader(path[0])
	}
	return os.Open(path[0])
}
