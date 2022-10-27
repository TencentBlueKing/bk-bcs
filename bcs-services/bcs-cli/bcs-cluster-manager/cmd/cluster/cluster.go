package cluster

import (
	"github.com/spf13/cobra"
)

func NewClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "cluster-related operations",
	}

	cmd.AddCommand(newCreateCmd())

	return cmd
}
