package federationcluster

import (
	"github.com/spf13/cobra"
)

var (
	clusterID           string
	federationClusterID string
)

func NewClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "federationcluster",
		Short: "federation cluster-related operations",
	}

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newAddCmd())

	return cmd
}
