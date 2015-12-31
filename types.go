package cfbackup

//BackupContext - stores the base context information for a backup/restore
type BackupContext struct {
	TargetDir       string
	S3Domain        string
	BucketName      string
	AccessKeyID     string
	SecretAccessKey string
	IsS3Archive     bool
}

// Tile is a deployable component that can be backed up
type Tile interface {
	Backup() error
	Restore() error
}

type connBucketInterface interface {
	Host() string
	AdminUser() string
	AdminPass() string
	OpsManagerUser() string
	OpsManagerPass() string
	Destination() string
}

type action func() error

type actionAdaptor func(t Tile) action
