package main

import (
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper"
	"github.com/cloudnative-pg/cnpg-i/pkg/backup"
	"github.com/cloudnative-pg/cnpg-i/pkg/lifecycle"
	"github.com/cloudnative-pg/cnpg-i/pkg/operator"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	backupImpl "github.com/dougkirkley/cnpg-plugin-s3-backup/internal/backup"
	"github.com/dougkirkley/cnpg-plugin-s3-backup/internal/identity"
	lifecycleImpl "github.com/dougkirkley/cnpg-plugin-s3-backup/internal/lifecycle"
	operatorImpl "github.com/dougkirkley/cnpg-plugin-s3-backup/internal/operator"
)

// newPluginCmd creates the `plugin` command
func newPluginCmd() *cobra.Command {
	cmd := pluginhelper.CreateMainCmd(identity.Identity{}, func(server *grpc.Server) error {
		operator.RegisterOperatorServer(server, operatorImpl.Operator{})
		backup.RegisterBackupServer(server, backupImpl.Server{})
		lifecycle.RegisterOperatorLifecycleServer(server, lifecycleImpl.Lifecycle{})
		return nil
	})

	cmd.Use = "plugin"
	cmd.Short = "Runs the cnpg-i plugin server for Cloudnative-PG backups to S3"

	return cmd
}
