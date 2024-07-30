package main

import (
	"log"

	"github.com/shaharia-lab/smarty-pants/backend/cmd"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "smarty-pants-ai",
		Version: Version,
	}

	rootCmd.AddCommand(cmd.NewStartCommand())

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("could not run the application: %v", err)
	}
}
