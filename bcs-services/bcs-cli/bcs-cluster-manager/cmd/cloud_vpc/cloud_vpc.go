package cloudvpc

import "github.com/spf13/cobra"

var (
	file    string
	cloudID string
	vpcID   string
)

func NewCloudVPCCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloudvpc",
		Short: "cloud vpc-related operations",
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newListCloudRegionsCmd())
	cmd.AddCommand(newGetVPCCidrCmd())

	return cmd
}
