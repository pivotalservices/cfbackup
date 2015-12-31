package cfbackup

func NewBackupContext(targetDir string, env map[string]string) (backupContext BackupContext) {
	backupContext = BackupContext{
		TargetDir: targetDir,
	}

	if useS3(env) {
		backupContext.AccessKeyID = env[AccessKeyIDVarname]
		backupContext.SecretAccessKey = env[SecretAccessKeyVarname]
		backupContext.BucketName = env[BucketNameVarname]
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
