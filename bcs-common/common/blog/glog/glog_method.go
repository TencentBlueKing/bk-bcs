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

package glog

import (
	"strconv"
	"sync"
)

func SetV(level Level) {
	logging.verbosity.Set(strconv.Itoa(int(level)))
}

var once sync.Once

// Init glog from commandline params
func InitLogs(toStderr, alsoToStderr bool, verbose int32, stdErrThreshold, vModule, traceLocation, dir string, maxSize uint64, maxNum int) {
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
