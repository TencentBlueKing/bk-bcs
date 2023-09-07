package main

import (
	"context"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vault-sidecar",
	Short: "vault server auto unseal, plugin register & ctl",
}

func execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(serverCmd())
}
