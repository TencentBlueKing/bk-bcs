package namespacequota

import "github.com/spf13/cobra"

var (
	file                string
	clusterID           string
	namespace           string
	federationClusterID string
	offset              uint32
	limit               uint32
)

func NewNamespaceQuotaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespace quota",
		Short: "namespace quota-related operations",
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newCreateNamespaceWithQuotaCmd())

	return cmd
}
