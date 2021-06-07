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
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/jaeger"
	"github.com/opentracing/opentracing-go/log"
)

func initTracing(t *testing.T, serviceName string) (io.Closer, error) {
	tracer, err := tracing.NewInitTracing(serviceName,
		tracing.TracerSwitch("on"),
		tracing.TracerType(tracing.Jaeger),
		tracing.ReportLog(true),
		tracing.SampleType(jaeger.SamplerTypeConst),
		tracing.SampleParameter(1))
	if err != nil {
		t.Fatal(err)
	}

	return tracer.Init()
}

func TestSpanPackage(t *testing.T) {
	closer, err := initTracing(t, "span-package")
	if err != nil {
		t.Fatal()
		return
	}

	if closer != nil {
		defer closer.Close()
	}

	span, ctx := StartSpanFromContext(context.Background(), "say-hello-2")
	defer span.Finish()

	SetSpanCommonTag(span, "hello-to", "evan")
	SetSpanLogFields(span,
		log.Event("string-format"),
		log.String("value", "lixin"),
	)

	formatString(ctx, "evan")
	printHello(ctx, "evan")
}

func formatString(ctx context.Context, helloTo string) string {
	span, ctx := StartSpanFromContext(ctx, "formatString")
	defer span.Finish()

	helloStr := fmt.Sprintf("Hello, %s!", helloTo)
	span.LogFields(
		log.String("event", "string-format"),
		log.String("value", helloStr),
	)

	printHello(ctx, helloTo)

	return helloStr
}

func printHello(ctx context.Context, helloStr string) {
	span, _ := StartSpanFromContext(ctx, "printHello")
	defer span.Finish()

	println(helloStr)
	span.LogKV("event", "println")
	err := fmt.Errorf("build error %s", "test")
	if err != nil {
		SetSpanLogTagError(span, err)
	}
}
