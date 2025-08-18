package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/version"
)

func main() {
	// 定义命令行参数
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "显示版本信息")
	flag.Parse()

	// 如果指定了 --version 参数，显示版本信息并退出
	if showVersion {
		fmt.Printf("BCS API Gateway Syncing\n")
		fmt.Printf("Version: %s\n", version.BcsVersion)
		fmt.Printf("GitHash: %s\n", version.BcsGitHash)
		fmt.Printf("BuildTime: %s\n", version.BcsBuildTime)
		fmt.Printf("GoVersion: %s\n", version.GoVersion)
		os.Exit(0)
	}

	// 如果没有指定 --version，显示使用说明
	fmt.Println("BCS API Gateway Syncing")
	fmt.Println("使用方法: ./bcs-api-gateway-syncing --version")
	os.Exit(1)
}
