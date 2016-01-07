package cfbackup

import (
	"io"
	"os"
	ospath "path"
	"strings"

	"github.com/pivotalservices/gtils/log"
	"github.com/pivotalservices/gtils/osutils"
	"github.com/pivotalservices/gtils/storage"
	"github.com/xchapter7x/lo"
)

// NewBackupContext initializes a BackupContext
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

// Writer for writing to a file or s3 bucket
func (s *BackupContext) Writer(path ...string) (io.WriteCloser, error) {
	if s.IsS3Archive {
		s3FilePath := strings.Join(path, "/")
		lo.G.Debug("BackupContext.Writer()", log.Data{"s3FilePath": s3FilePath})

		s3, err := storage.SafeCreateS3Bucket(s.S3Domain, s.BucketName, s.AccessKeyID, s.SecretAccessKey)
		if err != nil {
			return nil, err
		}
		lo.G.Debug("Created Bucket", log.Data{"s3Domain": s.S3Domain, "bucketName": s.BucketName, "accessKeyID": s.AccessKeyID, "secretAccessKey": s.SecretAccessKey})
		return s3.NewWriter(s3FilePath)
	}
	return osutils.SafeCreate(path...)
}

// Reader for reading from a file or s3 bucket
func (s *BackupContext) Reader(path ...string) (io.ReadCloser, error) {
	if s.IsS3Archive {
		s3FilePath := strings.Join(path, "/")
		lo.G.Debug("BackupContext.Reader()", log.Data{"s3FilePath": s3FilePath})
		s3, err := storage.SafeCreateS3Bucket(s.S3Domain, s.BucketName, s.AccessKeyID, s.SecretAccessKey)
		if err != nil {
			return nil, err
		}
		return s3.NewReader(s3FilePath)
	}
	filePath := ospath.Join(path...)
	return os.Open(filePath)
}
