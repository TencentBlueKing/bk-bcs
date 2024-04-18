package main

import (
	goflag "flag"
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/cmd/server"
)

func main() {
	command := server.NewServerCommand()
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "server starting failed %s\n", err.Error())
		os.Exit(1)
	}
}
