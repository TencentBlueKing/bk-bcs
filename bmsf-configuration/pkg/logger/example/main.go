package main

import (
	"time"

	"bk-bscp/pkg/logger"
)

func main() {
	logger.InitLogger(
		logger.LogConfig{
			LogDir:          "./log",
			LogMaxSize:      500,
			LogMaxNum:       5,
			ToStdErr:        false,
			AlsoToStdErr:    false,
			Verbosity:       5,
			StdErrThreshold: "2",
		},
	)
	defer logger.CloseLogs()

	for {
		logger.V(3).Info("V-info xxxxxxxx")
		time.Sleep(time.Second)

		logger.Info("Info xxxxxxxx")
		time.Sleep(time.Second)

		logger.Warn("Warn xxxxxxxx")
		time.Sleep(time.Second)

		logger.Error("Error xxxxxxxx")
		time.Sleep(time.Second)
	}
}
