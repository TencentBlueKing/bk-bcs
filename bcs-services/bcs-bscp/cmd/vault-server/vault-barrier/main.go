package main

import (
	"os"

	"github.com/hashicorp/vault/command"
)

func main() {
	args := []string{
		"server",
		"-dev",
		"-dev-root-token-id=root",
		"-dev-plugin-dir=./vault/plugins", // 指定插件目录
	}

	// 可以启动自定义命令
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}

	os.Exit(command.Run(args))
}
