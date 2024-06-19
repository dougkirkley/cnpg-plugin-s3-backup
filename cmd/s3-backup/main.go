// Package main is the entrypoint of the application
package main

import (
	"github.com/spf13/cobra"
	"log"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "s3-backup",
	}

	rootCmd.AddCommand(
		newPluginCmd(),
		newRestoreCmd(),
	)

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err.Error())
	}
}
