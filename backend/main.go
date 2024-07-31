package main

import (
	"log"

	"github.com/shaharia-lab/smarty-pants/backend/cmd"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "smarty-pants",
		Version: formatVersion(),
	}

	rootCmd.AddCommand(cmd.NewStartCommand())

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("could not run the application: %v", err)
	}
}

func formatVersion() string {
	return version + " (commit: " + commit + ", built at: " + date + ")"
}
