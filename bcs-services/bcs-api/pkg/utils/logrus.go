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

package utils

import (
	"fmt"
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// TraceContext xxx
type TraceContext struct {
	ShowFileno    bool
	ShowFuncField bool
}

// NewTraceContext xxx
func NewTraceContext(showFileno, showFuncField bool) *TraceContext {
	return &TraceContext{
		ShowFuncField: showFuncField,
		ShowFileno:    showFileno,
	}
}

// Levels xxx
func (hook TraceContext) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire adds filename and lino number info compare to the original log format
func (hook TraceContext) Fire(entry *logrus.Entry) error {
	pc := make([]uintptr, 3, 3)
	cnt := runtime.Callers(6, pc)

	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i] - 1)
		name := fu.Name()
		if !strings.Contains(name, "github.com/sirupsen/logrus") {
			file, line := fu.FileLine(pc[i] - 1)
			if hook.ShowFileno {
				// Prepend file and line number instead update fields
				entry.Message = fmt.Sprintf("[%s:%d] %s", path.Base(file), line, entry.Message)
			}
			if hook.ShowFuncField {
				entry.Data["func"] = path.Base(name)
			}
			break
		}
	}
	return nil
}

// SimpleContext xxx
type SimpleContext struct{}

// Levels xxx
func (hook SimpleContext) Levels() []logrus.Level {
	return []logrus.Level{logrus.DebugLevel, logrus.InfoLevel}
}

// Fire xxx
func (hook SimpleContext) Fire(entry *logrus.Entry) error {
	traceTypes := []string{"file", "func", "line"}
	for _, traceType := range traceTypes {
		if _, ok := entry.Data[traceType]; ok {
			delete(entry.Data, traceType)
		}
	}

	return nil
}

func initializeLogger(level logrus.Level) *logrus.Logger {
	hookedLogger := logrus.New()
	hookedLogger.AddHook(NewTraceContext(true, false))
	// hookedLogger.AddHook(&SimpleContext{})
	hookedLogger.SetLevel(level)
	// Log as JSON instead of the default ASCII formatter.
	hookedLogger.Formatter = &logrus.TextFormatter{FullTimestamp: true}

	return hookedLogger
}

// UpdateLoggerLevel xxx
func UpdateLoggerLevel(level logrus.Level) {
	logger.Level = level
}

// GetLogger xxx
func GetLogger() *logrus.Logger {
	return logger
}

var (
	logger *logrus.Logger
)

func init() {
	logger = initializeLogger(logrus.InfoLevel)
}
