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

package tracing

import (
	"context"
	"fmt"
	"io"
	"testing"

	opentrace "github.com/opentracing/opentracing-go"
	tracinglog "github.com/opentracing/opentracing-go/log"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/jaeger"
)

func initTracing(t *testing.T, serviceName string) (io.Closer, error) {
	tracer, err := NewInitTracing(
		serviceName,
		TracerSwitch("on"),
		TracerType(Jaeger),
		ReportLog(true),
		SampleType(jaeger.SamplerTypeConst),
		SampleParameter(1))
	if err != nil {
		t.Fatal(err)
	}

	return tracer.Init()
}

func TestNewInitTracing(t *testing.T) {
	closer, err := initTracing(t, "tracing-init")
	if err != nil {
		t.Fatal()
		return
	}

	if closer != nil {
		defer func(closer io.Closer) {
			_ = closer.Close()
		}(closer)
	}

	spanTest()
}

func spanTest() {
	span := opentrace.StartSpan("spanTest")
	span.SetTag("hello", "world")
	defer span.Finish()

	ctx := opentrace.ContextWithSpan(context.Background(), span)
	formatString(ctx, "evan")
	printString(ctx, "evan")
}

func formatString(ctx context.Context, helloTo string) {
	span, ctx := opentrace.StartSpanFromContext(ctx, "formatString")
	defer span.Finish()

	helloStr := fmt.Sprintf("hello, %s", helloTo)
	span.LogFields(
		tracinglog.String("event", "string-format"),
		tracinglog.String("value", helloStr),
	)

	printString(ctx, helloStr)
}

func printString(ctx context.Context, helloStr string) {
	span, _ := opentrace.StartSpanFromContext(ctx, "printString")
	defer span.Finish()

	println(helloStr)
	span.LogKV("event", "println")
}
