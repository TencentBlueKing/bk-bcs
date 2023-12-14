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

package logs

import (
	"bytes"
	"context"
	"fmt"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs/glog"
)

// logPrefix define access log service's log prefix.
type logPrefix string

const (
	redisPrefix logPrefix = "[redis] "
	grpcPrefix  logPrefix = "[grpc] "
)

// logger bscp logger, used to access other service logs.
type logger struct {
	// Prefix other service print log prefix.
	Prefix logPrefix
}

// newLogger new logger.
func newLogger(prefix logPrefix) *logger {
	return &logger{
		Prefix: prefix,
	}
}

// Info print client to the INFO log.
func (l *logger) Info(args ...interface{}) {
	if V(6) {
		Infof(string(l.Prefix)+"%s", convertToString(args...))
	}
}

// Infoln print client to the INFO log.
func (l *logger) Infoln(args ...interface{}) {
	if V(6) {
		Infof(string(l.Prefix)+"%s", convertToString(args...))
	}
}

// Infof print client info logs.
func (l *logger) Infof(format string, args ...interface{}) {
	if V(6) {
		Infof(string(l.Prefix)+format, args...)
	}
}

// Warning print client warning logs.
func (l *logger) Warning(args ...interface{}) {
	Warnf(string(l.Prefix)+"%s", convertToString(args...))
}

// Warningln print client warning logs.
func (l *logger) Warningln(args ...interface{}) {
	Warnf(string(l.Prefix)+"%s", convertToString(args...))
}

// Warningf print client warning logs.
func (l *logger) Warningf(format string, args ...interface{}) {
	Warnf(string(l.Prefix)+format, args...)
}

// Error print client error logs.
func (l *logger) Error(args ...interface{}) {
	Errorf(string(l.Prefix)+"%s", convertToString(args...))
}

// Errorln print client error logs.
func (l *logger) Errorln(args ...interface{}) {
	Errorf(string(l.Prefix)+"%s", convertToString(args...))
}

// Errorf print client error logs.
func (l *logger) Errorf(format string, args ...interface{}) {
	Errorf(string(l.Prefix)+format, args...)
}

// Fatal print client fatal logs.
func (l *logger) Fatal(args ...interface{}) {
	Errorf(string(l.Prefix)+"Fatal %s", convertToString(args...))
}

// Fatalln print client fatal logs.
func (l *logger) Fatalln(args ...interface{}) {
	Errorf(string(l.Prefix)+"Fatal %s", convertToString(args...))
}

// Fatalf print client fatal logs.
func (l *logger) Fatalf(format string, args ...interface{}) {
	Errorf(string(l.Prefix)+"Fatal "+format, args)
}

// Printf print logs.
func (l *logger) Printf(ctx context.Context, format string, v ...interface{}) {
	Errorf(string(l.Prefix)+format, v...)
}

// V reports whether verbosity at the call site is at least the requested level.
func (l *logger) V(v int) bool {
	return bool(V(glog.Level(v)))
}

// convertToString convert arguments in to string in the manner of fmt.Print
func convertToString(args ...interface{}) string {
	buf := new(bytes.Buffer)
	fmt.Fprintln(buf, args...)
	return buf.String()
}
