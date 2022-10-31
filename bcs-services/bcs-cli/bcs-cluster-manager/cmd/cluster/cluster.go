package cluster

import (
	"github.com/spf13/cobra"
)

var (
	file      string
	clusterID string
	nodes     []string
	limit     uint32
)

func NewClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "cluster-related operations",
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newRetryCmd())
	cmd.AddCommand(newAddNodesCmd())
	cmd.AddCommand(newDeleteNodesCmd())
	cmd.AddCommand(newCheckCloudKubeConfigCmd())
	cmd.AddCommand(newImportCmd())
	cmd.AddCommand(newListNodesCmd())
	cmd.AddCommand(newListCommonCmd())

	return cmd
}
