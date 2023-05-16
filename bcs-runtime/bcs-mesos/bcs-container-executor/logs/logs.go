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

// Package logs xxx
package logs

import (
	"log"
	"os"
	"sync"
)

var stdLogger *log.Logger
var errLogger *log.Logger
var once sync.Once

func init() {
	once.Do(func() {
		// init conLogger only once
		errLogger = log.New(os.Stderr, "", log.LstdFlags)
		stdLogger = log.New(os.Stdout, "", log.LstdFlags)
	})
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func Fatal(v ...interface{}) {
	errLogger.Fatal(v...)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func Fatalf(format string, v ...interface{}) {
	errLogger.Fatalf(format, v...)
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func Fatalln(v ...interface{}) {
	errLogger.Fatalln(v...)
}

// Info is equivalent to Print()
func Info(v ...interface{}) {
	stdLogger.Print(v...)
}

// Infof is equivalent to Printf()
func Infof(format string, v ...interface{}) {
	stdLogger.Printf(format, v...)
}

// Infoln is equivalent to Println()
func Infoln(v ...interface{}) {
	stdLogger.Println(v...)
}

// Error is equivalent to Print()
func Error(v ...interface{}) {
	stdLogger.Print(v...)
}

// Errorf is equivalent to Printf()
func Errorf(format string, v ...interface{}) {
	stdLogger.Printf(format, v...)
}

// Errorln is equivalent to Println()
func Errorln(v ...interface{}) {
	stdLogger.Println(v...)
}
