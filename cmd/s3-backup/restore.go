package main

import (
	"github.com/dougkirkley/cnpg-plugin-s3-backup/internal/backup/executor"
	"github.com/spf13/cobra"
	"os"
)

// newRestoreCmd creates the `restore` command
func newRestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restores a Postgres backup to the current Postgres cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			bucket := os.Getenv("AWS_BUCKET")
			prefix := os.Getenv("BACKUP_PREFIX")
			backupName, _ := cmd.Flags().GetString("backup-name")

			rep, err := executor.NewRepository(
				bucket,
				prefix,
			)
			if err != nil {
				return err
			}

			err = rep.Restore(cmd.Context(), backupName)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().String("backup-name", "", "The backup name to restore")

	return cmd
}
