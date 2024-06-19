package backup

import (
	"context"
	"github.com/cloudnative-pg/cnpg-i/pkg/backup"
	"github.com/dougkirkley/cnpg-plugin-s3-backup/internal/backup/executor"
	"os"
	"time"
)

// Server is the implementation of the identity service
type Server struct {
	backup.BackupServer
}

// GetCapabilities gets the capabilities of the Backup service
func (Server) GetCapabilities(
	context.Context,
	*backup.BackupCapabilitiesRequest,
) (*backup.BackupCapabilitiesResult, error) {
	return &backup.BackupCapabilitiesResult{
		Capabilities: []*backup.BackupCapability{
			{
				Type: &backup.BackupCapability_Rpc{
					Rpc: &backup.BackupCapability_RPC{
						Type: backup.BackupCapability_RPC_TYPE_BACKUP,
					},
				},
			},
		},
	}, nil
}

// Backup take a physical backup using Kopia
func (Server) Backup(
	ctx context.Context,
	request *backup.BackupRequest,
) (*backup.BackupResult, error) {
	bucket := os.Getenv("AWS_BUCKET")
	prefix := os.Getenv("BACKUP_PREFIX")

	return PerformBackup(ctx, bucket, prefix)
}

func PerformBackup(ctx context.Context, bucket string, prefix string) (*backup.BackupResult, error) {
	rep, err := executor.NewRepository(
		bucket,
		prefix,
	)
	if err != nil {
		return nil, err
	}

	exec := executor.NewS3Executor(
		rep,
	)

	startedAt := time.Now()
	backupInfo, err := exec.Backup(ctx)
	if err != nil {
		return nil, err
	}

	return &backup.BackupResult{
		BackupId:          backupInfo.BackupName,
		BackupName:        backupInfo.BackupName,
		StartedAt:         startedAt.Unix(),
		StoppedAt:         time.Now().Unix(),
		BeginWal:          exec.GetBeginWal(),
		EndWal:            exec.GetEndWal(),
		BeginLsn:          string(backupInfo.BeginLSN),
		EndLsn:            string(backupInfo.EndLSN),
		BackupLabelFile:   backupInfo.LabelFile,
		TablespaceMapFile: backupInfo.SpcmapFile,
		Online:            true,
	}, nil
}
