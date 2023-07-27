package main

import (
	"os"

	"github.com/hashicorp/vault/command"
)

func main() {

	os.Args = append(os.Args, "server")
	os.Args = append(os.Args, "-dev")
	os.Args = append(os.Args, "-dev-root-token-id=root")
	os.Args = append(os.Args, "-dev-plugin-dir=./vault/plugins") // 指定插件目录

	os.Exit(command.Run(os.Args[1:]))
}
