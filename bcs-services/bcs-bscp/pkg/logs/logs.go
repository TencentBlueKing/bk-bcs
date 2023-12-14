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
 */

// Package logs NOTES
package logs

import (
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/grpclog"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs/glog"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/version"
)

// GlogWriter serves as a bridge between the standard log package and the glog package.
type GlogWriter struct{}

// Write implements the io.Writer interface.
func (writer GlogWriter) Write(data []byte) (n int, err error) {
	glog.Info(string(data))
	return len(data), nil
}

var once sync.Once

// logger export func.
var (
	V = glog.V

	Infof      = glog.Infof
	InfoDepthf = glog.InfoDepthf

	Warnf = glog.Warningf

	Errorf      = glog.Errorf
	ErrorDepthf = glog.ErrorDepthf

	Fatalf = glog.Fatalf
)

// LogConfig is Log configuration.
type LogConfig struct {
	LogDir             string
	LogMaxSize         uint32
	LogLineMaxSize     uint32
	LogMaxNum          uint
	RestartNoScrolling bool
	ToStdErr           bool
	AlsoToStdErr       bool
	Verbosity          uint
	StdErrThreshold    string
	VModule            string
	TraceLocation      string
}

// InitLogger initializes logs the way we want for blog.
func InitLogger(logConfig LogConfig) {
	glog.InitLogs(logConfig.ToStdErr, logConfig.AlsoToStdErr, logConfig.RestartNoScrolling,
		int32(logConfig.Verbosity), logConfig.StdErrThreshold, logConfig.VModule, logConfig.TraceLocation,
		logConfig.LogDir, logConfig.LogMaxSize, logConfig.LogLineMaxSize, int(logConfig.LogMaxNum))

	// show inner start info.
	glog.Info(version.GetStartInfo())
	glog.Flush()

	once.Do(func() {
		log.SetOutput(GlogWriter{})
		log.SetFlags(0)

		// access other service log.
		redis.SetLogger(newLogger(redisPrefix))
		grpclog.SetLoggerV2(newLogger(grpcPrefix))

		// The default glog flush interval is 5 seconds, which is frighteningly long.
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for {
				<-ticker.C
				glog.Flush()
			}
		}()
	})
}

// SetV set the level of logger.
func SetV(level int32) {
	glog.SetV(glog.Level(level))
}

// GetV get the level of logger.
func GetV() int32 {
	return int32(glog.GetV())
}

// CloseLogs closes the logger.
func CloseLogs() {
	glog.Flush()
}
