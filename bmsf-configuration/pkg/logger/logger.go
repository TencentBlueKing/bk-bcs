/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package logger

import (
	"log"
	"sync"
	"time"

	"bk-bscp/pkg/logger/glog"
	"bk-bscp/pkg/version"
)

// GlogWriter serves as a bridge between the standard log package and the glog package.
type GlogWriter struct{}

// Write implements the io.Writer interface.
func (writer GlogWriter) Write(data []byte) (n int, err error) {
	glog.Info(string(data))
	return len(data), nil
}

var once sync.Once

// LogConfig is Log configuration.
type LogConfig struct {
	LogDir          string `json:"log_dir" value:"./logs" usage:"If non-empty, write log files in this directory" mapstructure:"log_dir"`
	LogMaxSize      uint64 `json:"log_max_size" value:"500" usage:"Max size (MB) per log file." mapstructure:"log_max_size"`
	LogMaxNum       int    `json:"log_max_num" value:"10" usage:"Max num of log file. The oldest will be removed if there is a extra file created." mapstructure:"log_max_num"`
	ToStdErr        bool   `json:"logtostderr" value:"false" usage:"log to standard error instead of files" mapstructure:"logtostderr"`
	AlsoToStdErr    bool   `json:"alsologtostderr" value:"false" usage:"log to standard error as well as files" mapstructure:"alsologtostderr"`
	Verbosity       int32  `json:"v" value:"0" usage:"log level for V logs, 3:debug 2:info 1:warn 0:error" mapstructure:"v"`
	StdErrThreshold string `json:"stderrthreshold" value:"2" usage:"logs at or above this threshold go to stderr" mapstructure:"stderrthreshold"`
	VModule         string `json:"vmodule" value:"" usage:"comma-separated list of pattern=N settings for file-filtered logging" mapstructure:"vmodule"`
	TraceLocation   string `json:"log_backtrace_at" value:"" usage:"when logging hits line file:N, emit a stack trace" mapstructure:"log_backtrace_at"`
}

// InitLogger initializes logs the way we want for blog.
func InitLogger(logConfig LogConfig) {
	glog.InitLogs(logConfig.ToStdErr, logConfig.AlsoToStdErr, logConfig.Verbosity, logConfig.StdErrThreshold,
		logConfig.VModule, logConfig.TraceLocation, logConfig.LogDir, logConfig.LogMaxSize, logConfig.LogMaxNum)

	// show inner start info.
	glog.Info(version.GetStartInfo())
	glog.Flush()

	once.Do(func() {
		log.SetOutput(GlogWriter{})
		log.SetFlags(0)

		// The default glog flush interval is 30 seconds, which is frighteningly long.
		go func() {
			d := time.Duration(5 * time.Second)
			tick := time.Tick(d)

			for {
				<-tick
				glog.Flush()
			}
		}()
	})
}

// SetV sets the level of logger.
func SetV(level int32) {
	glog.SetV(glog.Level(level))
}

// CloseLogs closes the logger.
func CloseLogs() {
	glog.Flush()
}

var (
	// V reports whether verbosity at the call site is at least the requested level.
	// The returned value is a boolean of type Verbose, which implements Info, Infoln
	// and Infof. These methods will write to the Info log if called.
	// Thus, one may write either
	//	if glog.V(2) { glog.Info("log this") }
	// or
	//	glog.V(2).Info("log this")
	// The second form is shorter but the first is cheaper if logging is off because it does
	// not evaluate its arguments.
	//
	// Whether an individual call to V generates a log record depends on the setting of
	// the -v and --vmodule flags; both are off by default. If the level in the call to
	// V is at least the value of -v, or of -vmodule for the source file containing the
	// call, the V call will log.
	V = glog.V

	// Info logs to the INFO log.
	// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
	Info = glog.Infof

	// Infof logs to the INFO log.
	// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
	Infof = glog.Infof

	// Warn logs to the WARNING and INFO logs.
	// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
	Warn = glog.Warningf

	// Warnf logs to the WARNING and INFO logs.
	// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
	Warnf = glog.Warningf

	// Error logs to the ERROR, WARNING, and INFO logs.
	// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
	Error = glog.Errorf

	// Errorf logs to the ERROR, WARNING, and INFO logs.
	// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
	Errorf = glog.Errorf

	// Fatal logs to the FATAL, ERROR, WARNING, and INFO logs,
	// including a stack trace of all running goroutines, then calls os.Exit(255).
	// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
	Fatal = glog.Fatalf

	// Fatalf logs to the FATAL, ERROR, WARNING, and INFO logs,
	// including a stack trace of all running goroutines, then calls os.Exit(255).
	// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
	Fatalf = glog.Fatalf
)
