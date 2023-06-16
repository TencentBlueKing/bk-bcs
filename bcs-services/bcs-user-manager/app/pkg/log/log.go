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

package log

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/google/uuid"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/utils"
)

// Logger define logger
type Logger struct {
	requestID string
}

// Log get logger with ctx
func Log(ctx context.Context) *Logger {
	if ctx == nil {
		return &Logger{requestID: uuid.New().String()}
	}
	if id, ok := ctx.Value(utils.ContextValueKeyRequestID).(string); ok {
		return &Logger{requestID: id}
	}
	return &Logger{requestID: uuid.New().String()}
}

func (l *Logger) getPrefix() string {
	return fmt.Sprintf("[%s] ", l.requestID)
}

// Info log Info
func (l *Logger) Info(args ...interface{}) {
	l.Infof("", args)
}

// Infof log Infof
func (l *Logger) Infof(format string, args ...interface{}) {
	blog.Infof(l.getPrefix()+format, args...)
}

// Warn log Warn
func (l *Logger) Warn(args ...interface{}) {
	l.Warnf("", args)
}

// Warnf log Warnf
func (l *Logger) Warnf(format string, args ...interface{}) {
	blog.Warnf(l.getPrefix()+format, args...)
}

// Error log Error
func (l *Logger) Error(format string, args ...interface{}) {
	l.Errorf(format, args)
}

// Errorf log Errorf
func (l *Logger) Errorf(format string, args ...interface{}) {
	blog.Errorf(l.getPrefix()+format, args...)
}

// Fatal log Fatal
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.Fatalf(format, args)
}

// Fatalf log Fatalf
func (l *Logger) Fatalf(format string, args ...interface{}) {
	blog.Fatalf(l.getPrefix()+format, args...)
}
