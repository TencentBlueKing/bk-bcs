package node

import "github.com/spf13/cobra"

var (
	innerIP     string
	innerIPs    []string
	status      string
	nodeGroupID string
	clusterID   string
)

func NewNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "node-related operations",
	}

	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newCheckNodeInClusterCmd())
	cmd.AddCommand(newCordonCmd())
	cmd.AddCommand(newUnCordonCmd())
	cmd.AddCommand(newDrainCmd())

	return cmd
}
