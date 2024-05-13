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

// Go support for leveled logs, analogous to https://code.google.com/p/google-glog/
//
// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package glog

import (
	"strconv"
	"sync"
)

// SetV set the level of logging.
func SetV(level Level) {
	logging.verbosity.Set(strconv.Itoa(int(level))) // nolint error not checked
}

// GetV get the level of logging.
func GetV() Level {
	return logging.verbosity
}

var once sync.Once

// InitLogs inits glog from commandline params.
func InitLogs(toStderr, alsoToStderr, restartScrolling bool, verbose int32, stdErrThreshold,
	vModule, traceLocation, dir string, maxSize, maxLineSize uint32, maxNum int) {
	once.Do(func() {
		logging.toStderr = toStderr
		logging.alsoToStderr = alsoToStderr
		logging.verbosity.Set(strconv.Itoa(int(verbose))) // nolint error not checked
		logging.stderrThreshold.Set(stdErrThreshold)      // nolint error not checked
		logging.vmodule.Set(vModule)                      // nolint error not checked
		logging.traceLocation.Set(traceLocation)          // nolint error not checked

		logNoScrolling = restartScrolling
		lineMaxSize = maxLineSize * 1024
		logMaxNum = maxNum
		logMaxSize = maxSize * 1024 * 1024
		logDir = dir
	})
}
