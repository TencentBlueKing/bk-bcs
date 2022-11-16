/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package logs

import (
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	InitLogger(
		LogConfig{
			LogDir:             "./log",
			LogLineMaxSize:     2,
			LogMaxSize:         500,
			LogMaxNum:          5,
			RestartNoScrolling: false,
			ToStdErr:           false,
			AlsoToStdErr:       false,
			Verbosity:          5,
			StdErrThreshold:    "2",
		},
	)
	defer CloseLogs()

	for {
		intervalTime := time.Second

		logContent := ""
		for i := 0; i < 3*1024; i++ {
			logContent += "#"
		}
		Infof("log line max size test: %s", logContent)
		time.Sleep(intervalTime)

		V(3).Info("V-info xxxxxxxx")
		time.Sleep(intervalTime)

		Infof("Infof xxxxxxxx")
		time.Sleep(intervalTime)

		Warnf("Warnf xxxxxxxx")
		time.Sleep(intervalTime)

		Errorf("Errorf xxxxxxxx")
		time.Sleep(intervalTime)

		InfoDepthf(1, "InfofDepthf xxxxxxxx")
		time.Sleep(intervalTime)

		ErrorDepthf(1, "ErrorfDepthf xxxxxxxx")
		time.Sleep(intervalTime)
	}
}
