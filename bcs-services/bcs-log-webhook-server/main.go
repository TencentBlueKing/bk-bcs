package main

import (
	"fmt"
	"os"
	"runtime"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-log-webhook-server/app"
	"bk-bcs/bcs-services/bcs-log-webhook-server/options"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	op := options.NewServerOption()
	if err := options.Parse(op); err != nil {
		fmt.Printf("parse options failed: %v\n", err)
		os.Exit(1)
	}

	blog.InitLogs(op.LogConfig)
	defer blog.CloseLogs()

	app.Run(op)
}
