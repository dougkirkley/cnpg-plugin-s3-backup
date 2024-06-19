package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudnative-pg/cloudnative-pg/pkg/management/url"
	"github.com/cloudnative-pg/cloudnative-pg/pkg/utils"
	"github.com/google/uuid"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/cloudnative-pg/cloudnative-pg/pkg/management/postgres/webserver"
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/logging"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

const podIP = "127.0.0.1"

var (
	ErrBackupNotStarted = errors.New("backup not started")
	ErrBackupNotStopped = errors.New("backup not stopped")
)

var backupModeBackoff = wait.Backoff{
	Steps:    10,
	Duration: 1 * time.Second,
	Factor:   5.0,
	Jitter:   0.1,
}

// Executor manages the execution of a backup
type Executor struct {
	backupClient         webserver.BackupClient
	beginWal             string
	endWal               string
	backup               string
	repository           *Repository
	backupClientEndpoint string
	executed             bool
}

// GetBeginWal returns the beginWal value, panics if the executor was not executed
func (executor *Executor) GetBeginWal() string {
	if !executor.executed {
		panic("beginWal: please run take backup before trying to access this value")
	}
	return executor.beginWal
}

// GetEndWal returns the endWal value, panics if the executor was not executed
func (executor *Executor) GetEndWal() string {
	if !executor.executed {
		panic("endWal: please run take backup before trying to access this value")
	}
	return executor.endWal
}

// newExecutor creates a new backup Executor
func newExecutor(repo *Repository, endpoint string) *Executor {
	backupName, _ := uuid.NewUUID()
	return &Executor{
		backupClient:         webserver.NewBackupClient(),
		backup:               backupName.String(),
		repository:           repo,
		backupClientEndpoint: endpoint,
	}
}

// NewS3Executor creates a new backup Executor
func NewS3Executor(repo *Repository) *Executor {
	return newExecutor(repo, podIP)
}

// Backup executes a backup. Returns the result and any error encountered
func (executor *Executor) Backup(ctx context.Context) (*webserver.BackupResultData, error) {
	defer func() {
		executor.executed = true
	}()

	contextLogger := logging.FromContext(ctx)
	contextLogger.Info("Preparing physical backup")
	if err := executor.setBackupMode(ctx); err != nil {
		return nil, err
	}

	contextLogger.Info("Copying files")
	if err := executor.execSnapshot(ctx); err != nil {
		return nil, err
	}

	contextLogger.Info("Finishing backup")
	return executor.unsetBackupMode(ctx)
}

// setBackupMode starts a backup by setting PostgreSQL in backup mode
func (executor *Executor) setBackupMode(ctx context.Context) error {
	logger := logging.FromContext(ctx)

	var currentWALErr error
	executor.beginWal, currentWALErr = executor.getCurrentWALFile(ctx)
	if currentWALErr != nil {
		return currentWALErr
	}

	if err := executor.backupClient.Start(ctx, executor.backupClientEndpoint, webserver.StartBackupRequest{
		ImmediateCheckpoint: true,
		WaitForArchive:      true,
		BackupName:          executor.backup,
		Force:               true,
	}); err != nil {
		logger.Error(err, "while requesting new backup on PostgreSQL")
		return err
	}

	logger.Info("Requesting PostgreSQL Backup mode")
	if err := retry.OnError(backupModeBackoff, retryOnBackupNotStarted, func() error {
		response, err := executor.backupClient.StatusWithErrors(ctx, executor.backupClientEndpoint)
		if err != nil {
			return err
		}

		if response.Data.Phase != webserver.Started {
			logger.V(4).Info("Backup still not started", "status", response.Data)
			return ErrBackupNotStarted
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("Backup Mode started")
	return nil
}

// execSnapshot  runs pg_dumpall against the postgres cluster
func (executor *Executor) execSnapshot(ctx context.Context) error {
	logger := logging.FromContext(ctx)

	logger.Info("Running pg_dumpall")
	err := executor.repository.Snapshot(ctx)
	if err != nil {
		return err
	}

	return nil
}

// unsetBackupMode stops a backup and resume PostgreSQL normal operation
func (executor *Executor) unsetBackupMode(ctx context.Context) (*webserver.BackupResultData, error) {
	logger := logging.FromContext(ctx)

	if err := executor.backupClient.Stop(ctx, executor.backupClientEndpoint, webserver.StopBackupRequest{
		BackupName: executor.backup,
	}); err != nil {
		logger.Error(err, "while requesting new backup on PostgreSQL")
		return nil, err
	}

	logger.Info("Stopping PostgreSQL Backup mode")
	var backupStatus webserver.BackupResultData
	if err := retry.OnError(backupModeBackoff, retryOnBackupNotStopped, func() error {
		response, err := executor.backupClient.StatusWithErrors(ctx, executor.backupClientEndpoint)
		if err != nil {
			return err
		}

		if response.Data.Phase != webserver.Completed {
			logger.V(4).Info("backup still not stopped", "status", response.Data)
			return ErrBackupNotStopped
		}

		backupStatus = *response.Data

		return nil
	}); err != nil {
		return nil, err
	}
	logger.Info("PostgreSQL Backup mode stopped")

	var err error
	executor.endWal, err = executor.getCurrentWALFile(ctx)
	if err != nil {
		return nil, err
	}

	return &backupStatus, nil
}

func retryOnBackupNotStarted(e error) bool {
	return errors.Is(e, ErrBackupNotStarted)
}

func retryOnBackupNotStopped(e error) bool {
	return errors.Is(e, ErrBackupNotStopped)
}

func (executor *Executor) getCurrentWALFile(ctx context.Context) (string, error) {
	const currentWALFileControlFile = "Latest checkpoint's REDO WAL file"

	controlDataOutput, err := getPgControlData(ctx)
	if err != nil {
		return "", err
	}

	return controlDataOutput[currentWALFileControlFile], nil
}

// getPgControlData obtains the pg_controldata from the instance HTTP endpoint
func getPgControlData(
	ctx context.Context,
) (map[string]string, error) {
	contextLogger := logging.FromContext(ctx)

	const (
		connectionTimeout = 2 * time.Second
		requestTimeout    = 30 * time.Second
	)

	// We want a connection timeout to prevent waiting for the default
	// TCP connection timeout (30 seconds) on lost SYN packets
	timeoutClient := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: connectionTimeout,
			}).DialContext,
		},
		Timeout: requestTimeout,
	}

	httpURL := url.Build(podIP, url.PathPGControlData, url.StatusPort)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, httpURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := timeoutClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			contextLogger.Error(err, "while closing body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		contextLogger.Info("Error while querying the pg_controldata endpoint",
			"statusCode", resp.StatusCode,
			"body", string(body))
		return nil, fmt.Errorf("error while querying the pg_controldata endpoint: %d", resp.StatusCode)
	}

	type pgControldataResponse struct {
		Data  string `json:"data,omitempty"`
		Error error  `json:"error,omitempty"`
	}

	var result pgControldataResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		result.Error = err
		return nil, err
	}

	return utils.ParsePgControldataOutput(result.Data), result.Error
}
