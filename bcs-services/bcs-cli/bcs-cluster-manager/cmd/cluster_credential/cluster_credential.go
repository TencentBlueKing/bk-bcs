package clustercredential

import "github.com/spf13/cobra"

var (
	file      string
	serverKey string
	offset    uint32
	limit     uint32
)

func NewClusterCredentialCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespace",
		Short: "namespace-related operations",
	}

	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newListCmd())

	return cmd
}
