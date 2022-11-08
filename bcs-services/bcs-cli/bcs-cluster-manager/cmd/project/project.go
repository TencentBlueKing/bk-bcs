package project

import "github.com/spf13/cobra"

var (
	file      string
	projectID string
)

func NewNamespaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "project-related operations",
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newListCmd())

	return cmd
}
