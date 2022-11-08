package namespace

import "github.com/spf13/cobra"

var (
	file                string
	name                string
	federationClusterID string
	projectID           string
	businessID          string
	isForced            bool
	offset              uint32
	limit               uint32
)

func NewNamespaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespace",
		Short: "namespace-related operations",
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newListCmd())

	return cmd
}
