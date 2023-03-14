package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"bscp.io/cmd/data-service/app"
	"bscp.io/pkg/cc"
	"bscp.io/pkg/runtime/flags"
)

// SysOpt is the system option
var SysOpt *cc.SysOption

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bk-bscp-dataservice",
	Short: "BSCP DataService",
	Run: func(cmd *cobra.Command, args []string) {
		app.RunServer(SysOpt)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// add system flags
	SysOpt = flags.SysFlags(pflag.CommandLine)
	rootCmd.Flags().AddFlagSet(pflag.CommandLine)

	cc.InitService(cc.DataServiceName)
}
