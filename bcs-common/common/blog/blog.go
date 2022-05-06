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

package blog

import (
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog/glog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// GlogWriter serves as a bridge between the standard log package and the glog package.
type GlogWriter struct{}

// Write implements the io.Writer interface.
func (writer GlogWriter) Write(data []byte) (n int, err error) {
	glog.Info(string(data))
	return len(data), nil
}

var once sync.Once

// InitLogs initializes logs the way we want for blog.
func InitLogs(logConfig conf.LogConfig) {
	glog.InitLogs(logConfig.ToStdErr,
		logConfig.AlsoToStdErr,
		logConfig.Verbosity,
		logConfig.StdErrThreshold,
		logConfig.VModule,
		logConfig.TraceLocation,
		logConfig.LogDir,
		logConfig.LogMaxSize,
		logConfig.LogMaxNum,
	)
	once.Do(func() {
		log.SetOutput(GlogWriter{})
		log.SetFlags(0)
		// The default glog flush interval is 30 seconds, which is frighteningly long.
		go func() {
			d := time.Duration(5 * time.Second)
			tick := time.Tick(d)
			for {
				select {
				case <-tick:
					glog.Flush()
				}
			}
		}()
	})
}

func CloseLogs() {
	glog.Flush()
}

var (
	Info      = glog.Infof
	Infof     = glog.Infof
	InfoDepth = glog.InfoDepth

	Warn      = glog.Warningf
	Warnf     = glog.Warningf
	WarnDepth = glog.WarningDepth

	Error      = glog.Errorf
	Errorf     = glog.Errorf
	ErrorDepth = glog.ErrorDepth

	Fatal  = glog.Fatal
	Fatalf = glog.Fatalf

	V = glog.V
)

func Debug(args ...interface{}) {
	if format, ok := (args[0]).(string); ok {
		glog.V(3).Infof(format, args[1:]...)
	} else {
		glog.V(3).Info(args...)
	}
}

func SetV(level int32) {
	glog.SetV(glog.Level(level))
}

// defaultRe and defaultHandler is for bcs-dns wrap its extra time tag in log.
// the extra time tag of bcs-dns: [04/Jan/2018:09:44:27 +0800]
var defaultRe = regexp.MustCompile(`\[\d{2}/\w+/\d{4}:\d{2}:\d{2}:\d{2} \+\d{4}\] `)
var defaultHandler WrapFunc = func(format string, args ...interface{}) string {
	src := fmt.Sprintf(format, args...)
	return defaultRe.ReplaceAllString(src, "")
}

// WrapFunc take the param the same as glog.Infof, and return string.
type WrapFunc func(string, ...interface{}) string

// Wrapper use WrapFunc to handle the log message before send it to glog.
// Can be use as:
//      var handler blog.WrapFunc = func(format string, args ...interface{}) string {
//          src := fmt.Sprintf(format, args...)
//          dst := regexp.MustCompile("boy").ReplaceAllString(src, "man")
//      }
//      blog.Wrapper(handler).V(2).Info("hello boy")
// And it will flush as:
//      I0104 09:44:27.796409   16233 blog.go:21] hello man
type Wrapper struct {
	Handler WrapFunc
	verbose glog.Verbose
}

// Info implementation
func (w *Wrapper) Info(format string, args ...interface{}) {
	if w.verbose {
		Info(w.Handler(format, args...))
	}
}

// Warn implementation
func (w *Wrapper) Warn(format string, args ...interface{}) {
	if w.verbose {
		Warn(w.Handler(format, args...))
	}
}

// Error implementation
func (w *Wrapper) Error(format string, args ...interface{}) {
	if w.verbose {
		Error(w.Handler(format, args...))
	}
}

// Fatal implementation
func (w *Wrapper) Fatal(format string, args ...interface{}) {
	if w.verbose {
		Fatal(w.Handler(format, args...))
	}
}

// V implementation
func (w *Wrapper) V(level glog.Level) *Wrapper {
	w.verbose = V(level)
	return w
}

// Wrap Wrapper function
func Wrap(handler WrapFunc) *Wrapper {
	if handler == nil {
		handler = defaultHandler
	}
	return &Wrapper{verbose: true, Handler: handler}
}
