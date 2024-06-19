package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/dougkirkley/cnpg-plugin-s3-backup/internal/archiver"
	"github.com/go-logr/logr"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/logging"
)

const (
	PGDumpall        = "pg_dumpall"
	Psql             = "psql"
	BackupTimeFormat = "20060102150405"
	workingDir       = "/backup"
)

// Repository represents a backup repository where
// base directories are stored
type Repository struct {
	bucket string
	path   string
	cfg    aws.Config
}

// NewRepository creates a new repository ensuring
// that the repository is initialized and ready to
// accept backups
func NewRepository(bucket string, path string) (*Repository, error) {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	params := &s3.HeadBucketInput{
		Bucket: &bucket,
	}
	if _, err = client.HeadBucket(context.TODO(), params); err != nil {
		var noBucket *types.NoSuchBucket
		if errors.Is(err, noBucket) {
			return nil, fmt.Errorf("bucket %s: not found", bucket)
		}
	}

	return &Repository{
		bucket: bucket,
		path:   path,
		cfg:    cfg,
	}, nil
}

// Snapshot takes a Snapshot of the Postgres cluster
func (repo *Repository) Snapshot(ctx context.Context) error {
	logger := logging.FromContext(ctx)

	client := s3.NewFromConfig(repo.cfg)

	logger.Info("Creating snapshot")
	file, err := executeBackup(ctx)
	if err != nil {
		return err
	}

	logger.Info("Archiving snapshot")
	file, err = archiveBackup(file)
	if err != nil {
		return err
	}

	key := filepath.Join(repo.path, file)
	fileName := filepath.Join(workingDir, file)

	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	input := &s3.PutObjectInput{
		Bucket: &repo.bucket,
		Key:    &key,
		Body:   f,
	}

	if err := os.Remove(fileName); err != nil {
		return err
	}

	logger.Info("Uploading snapshot")

	logger.Info(fmt.Sprintf("uploading key: %s, file: %s", key, filepath.Join(workingDir, file)))
	if _, err := client.PutObject(ctx, input); err != nil {
		logger.Error(err, fmt.Sprintf("Unable to upload object to remote bucket: key %s, error: %s", key, err.Error()))
		return err
	}

	return nil
}

func (repo *Repository) Restore(ctx context.Context, backupName string) error {
	logger := logging.FromContext(ctx)

	logger.Info("Restoring snapshot")

	logger.Info("Downloading snapshot")
	backupFilename, err := repo.downloadBackup(ctx, logger, backupName)
	if err != nil {
		return err
	}

	backupFile := filepath.Base(backupFilename)
	folderName := backupFile[:len(backupFile)-len(".tar.gz")]

	logger.Info("Extracting snapshot")
	if err := archiver.ExtractArchive(backupFilename, filepath.Join(workingDir)); err != nil {
		return err
	}

	logger.Info("Executing restore")
	if err := executeRestore(ctx, filepath.Join(workingDir, folderName)); err != nil {
		return err
	}

	if err := os.Remove(backupFilename); err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(workingDir, folderName)); err != nil {
		return err
	}

	return nil
}

func (repo *Repository) downloadBackup(ctx context.Context, logger logr.Logger, backupName string) (string, error) {
	client := s3.NewFromConfig(repo.cfg)

	input := &s3.GetObjectInput{
		Bucket: &repo.bucket,
		Key:    &backupName,
	}
	resp, err := client.GetObject(ctx, input)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Unable to download object from remote bucket: key %s, error: %s", backupName, err.Error()))
		return "", err
	}
	defer resp.Body.Close()

	backupFile := filepath.Join(workingDir, filepath.Base(backupName))
	fo, err := os.Create(backupFile)
	if err != nil {
		return "", err
	}
	defer fo.Close()

	if _, err := io.Copy(fo, resp.Body); err != nil {
		return "", err
	}

	return backupFile, nil
}

// executeBackup executes pg_dump against the app database
func executeBackup(ctx context.Context) (string, error) {
	now := time.Now()
	file := fmt.Sprintf("%s.sql", now.Format(BackupTimeFormat))

	args := []string{
		"-h",
		"/controller/run",
		"-f",
		filepath.Join(workingDir, file),
	}

	if err := executeCommand(ctx, PGDumpall, args...); err != nil {
		return "", err
	}

	return file, nil
}

// executeBackup executes psql -f  against the Postgres dump file
func executeRestore(ctx context.Context, backupFolder string) error {
	args := []string{
		"-h",
		"/controller/run",
		"-f",
		backupFolder,
	}
	if err := executeCommand(ctx, Psql, args...); err != nil {
		return err
	}

	return nil
}

// archiveBackup converts a file into a gunzipped tar archive
func archiveBackup(file string) (string, error) {
	srcFile := filepath.Join(workingDir, file)
	destFilename := fmt.Sprintf("%s.tar.gz", file)

	err := archiver.CreateArchive(
		filepath.Join(workingDir, destFilename),
		[]string{srcFile},
	)
	if err != nil {
		return "", err
	}

	if err := os.RemoveAll(srcFile); err != nil {
		return "", err
	}

	return destFilename, nil
}

// executeCommand executes a standard command
func executeCommand(ctx context.Context, command string, args ...string) error {
	var stdout, stderr bytes.Buffer
	logger := logging.FromContext(ctx)
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		logger.Error(err, "command failed", command, "args", args, "stdout", stdout.String(), "stderr", stderr.String())
		return err
	}

	logger.Info("command succeeded", command, "args", args, "stdout", stdout.String(), "stderr", stderr.String())
	return nil
}
