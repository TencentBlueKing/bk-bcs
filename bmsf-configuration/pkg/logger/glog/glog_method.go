package glog

import (
	"strconv"
	"sync"
)

// SetV sets the level of logging.
func SetV(level Level) {
	logging.verbosity.Set(strconv.Itoa(int(level)))
}

var once sync.Once

// InitLogs inits glog from commandline params.
func InitLogs(toStderr, alsoToStderr bool, verbose int32, stdErrThreshold,
	vModule, traceLocation, dir string, maxSize uint64, maxNum int) {
	once.Do(func() {
		logging.toStderr = toStderr
		logging.alsoToStderr = alsoToStderr
		logging.verbosity.Set(strconv.Itoa(int(verbose)))
		logging.stderrThreshold.Set(stdErrThreshold)
		logging.vmodule.Set(vModule)
		logging.traceLocation.Set(traceLocation)

		logMaxNum = maxNum
		logMaxSize = maxSize * 1024 * 1024
		logDir = dir
	})
}
