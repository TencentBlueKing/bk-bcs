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

// Package logger istead of go-machinery logger
package logger

import (
	"fmt"

	"github.com/RichardKnop/logging"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

var _ logging.LoggerInterface = &TaskLogger{}

// NewTaskLogger create a task logger
func NewTaskLogger() *TaskLogger {
	return &TaskLogger{}
}

// TaskLogger task logger
type TaskLogger struct{}

// Print print log
func (t TaskLogger) Print(args ...interface{}) {
	blog.Info(args...)
}

// Printf print log
func (t TaskLogger) Printf(format string, args ...interface{}) {
	blog.Infof(format, args...)
}

// Println print log
func (t TaskLogger) Println(args ...interface{}) {
	blog.Info(args...)
}

// Fatal fatal log
func (t TaskLogger) Fatal(args ...interface{}) {
	blog.Fatal(args...)
}

// Fatalf fatal log
func (t TaskLogger) Fatalf(format string, args ...interface{}) {
	blog.Fatalf(format, args...)
}

// Fatalln fatal log
func (t TaskLogger) Fatalln(args ...interface{}) {
	blog.Fatal(args...)
}

// Panic panic log
func (t TaskLogger) Panic(args ...interface{}) {
	e := fmt.Sprint(args...)
	blog.Error(e)
	panic(e)
}

// Panicf panic log
func (t TaskLogger) Panicf(format string, args ...interface{}) {
	e := fmt.Sprintf(format, args...)
	blog.Error(e)
	panic(e)
}

// Panicln panic log
func (t TaskLogger) Panicln(args ...interface{}) {
	e := fmt.Sprint(args...)
	blog.Error(e)
	panic(e)
}
