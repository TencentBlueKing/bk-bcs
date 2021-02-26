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
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog/glog"

	"github.com/google/uuid"
	grpcmeta "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	stackDepth      = 2
	defaultTraceID  = "qazsdfrw234gjkdj"
	defaultClientIP = "0.0.0.0"
)

// TraceHandlerKey struct
type TraceHandlerKey struct{}

// GetTraceFromContext get the Trace var from the context, if there is no such a trace utility, return nil
func GetTraceFromContext(ctx context.Context) Trace {
	if tracer, ok := ctx.Value(TraceHandlerKey{}).(Trace); ok {
		return tracer
	}

	return NewTraceInfo(defaultTraceID, "default-trace")
}

// WithTraceForContext will return a new context wrapped a trace handler around the original ctx
func WithTraceForContext(ctx context.Context, traceName string, traceID string) (context.Context, Trace) {
	tracer := NewTraceInfo(traceID, traceName)
	return context.WithValue(ctx, TraceHandlerKey{}, tracer), tracer
}

// Trace log trace based on blog, used to trace a http request
type Trace interface {
	// Info will print the args as the info level log
	Info(args ...interface{})
	// Infof will print the args with a format as the info level log
	Infof(format string, args ...interface{})
	// Warn will print the args as the warn level log
	Warn(args ...interface{})
	// Warnf will print the args with a format as the warn level log
	Warnf(format string, args ...interface{})
	// Error will print the args as the error level log
	Error(args ...interface{})
	// Errorf will print the args with a format as the error level log
	Errorf(format string, args ...interface{})
	// DefaultRequestInEvent will print request in event
	DefaultRequestInEvent(clientIP, schema, method string)
	// DefaultRequestOutEvent will print request out event
	DefaultRequestOutEvent()
}

type trace struct {
	startTime   time.Time
	traceID     string
	clientIP    string
	handlerName string
}

// NewTraceInfo create Trace by ID,handlerName
func NewTraceInfo(traceID, handlerName string) Trace {
	return &trace{
		traceID:     traceID,
		handlerName: handlerName,
		startTime:   time.Now(),
	}
}

// DefaultRequestInEvent for http request originInfo
func (t *trace) DefaultRequestInEvent(clientIP, schema, method string) {
	if t == nil {
		return
	}
	t.Infof("event=[request-in] clientIP=[%s] schema=[%s] method=[%s]", clientIP, schema, method)
}

// DefaultRequestOutEvent quit http request
func (t *trace) DefaultRequestOutEvent() {
	t.Infof("event=[request-out]")
}

func (t *trace) traceBody() string {
	if t == nil {
		return ""
	}
	var buffer bytes.Buffer

	buffer.WriteString("tname=[")
	buffer.WriteString(t.handlerName)
	buffer.WriteString("] ")

	buffer.WriteString("tid=[")
	buffer.WriteString(t.traceID)
	buffer.WriteString("] ")

	buffer.WriteString("tduration=[")

	return buffer.String()
}

func (t *trace) duration() time.Duration {
	if t == nil {
		return 0
	}
	return time.Since(t.startTime) / time.Millisecond
}

func (t *trace) log(out func(depth int, args ...interface{}), args ...interface{}) {
	if t == nil {
		return
	}

	var newArgs []interface{}
	newArgs = append(newArgs, t.traceBody())
	if len(args) > 0 {
		newArgs = append(newArgs, args...)
	}

	out(stackDepth, newArgs...)
}

func (t *trace) logf(out func(depth int, args ...interface{}), format string, args ...interface{}) {
	if t == nil {
		return
	}

	log := fmt.Sprintf(t.traceBody()+format, args...)
	out(stackDepth, log)
}

// Info print args
func (t *trace) Info(args ...interface{}) {
	t.log(glog.InfoDepth, args...)
}

// Infof print args with format
func (t *trace) Infof(format string, args ...interface{}) {
	t.logf(glog.InfoDepth, format, args...)
}

// Warn print args
func (t *trace) Warn(args ...interface{}) {
	t.log(glog.WarningDepth, args...)
}

// Warnf print args with format
func (t *trace) Warnf(format string, args ...interface{}) {
	t.logf(glog.WarningDepth, format, args...)
}

// Error print args
func (t *trace) Error(args ...interface{}) {
	t.log(glog.ErrorDepth, args...)
}

// Errorf print args with format
func (t *trace) Errorf(format string, args ...interface{}) {
	t.logf(glog.ErrorDepth, format, args...)
}

// RequestHeaderInfo headerInfo
type RequestHeaderInfo struct {
	TraceID  string
	ClientIP string
}

// GetRequestHeaderInfo get http request headerInfo
func GetRequestHeaderInfo(ctx context.Context) (*RequestHeaderInfo, error) {
	traceID, err := requestTraceIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	clientIP, err := getClientIPFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &RequestHeaderInfo{
		TraceID:  traceID,
		ClientIP: clientIP,
	}, nil
}

// requestTraceIDFromContext
func requestTraceIDFromContext(ctx context.Context) (string, error) {
	meta, ok := grpcmeta.FromIncomingContext(ctx)
	if !ok {
		errMsg := fmt.Errorf("get grpc metadata from context failed")
		return "", errMsg
	}

	requestIDs := meta.Get("X-Request-Id")
	if len(requestIDs) == 0 {
		// generate traceID
		return uuid.New().String(), nil
	}

	return requestIDs[0], nil
}

// GetClientIPFromContext
func getClientIPFromContext(ctx context.Context) (string, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("[getClinetIP] invoke FromContext() failed")
	}
	if pr.Addr == net.Addr(nil) {
		blog.Warnf("[getClientIP] peer.Addr is nil")
		return "", nil
	}

	addSlice := strings.Split(pr.Addr.String(), ":")
	if addSlice[0] == "[" {
		return "localhost", nil
	}

	return addSlice[0], nil
}
