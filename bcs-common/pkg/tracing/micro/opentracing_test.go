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

package micro

import (
	"context"
	"fmt"
	tracinglog "github.com/opentracing/opentracing-go/log"
	"io"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/jaeger"
	proto "github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/micro/proto"

	client "github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/registry/memory"
	microsvc "github.com/micro/go-micro/v2/service"
	grpcsvc "github.com/micro/go-micro/v2/service/grpc"
	"github.com/opentracing/opentracing-go"
)

func initTracing(t *testing.T, serviceName string) (io.Closer, error) {
	tracer, err := tracing.NewInitTracing(
		serviceName,
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

func TestNewHandlerWrapper(t *testing.T) {
	closer, err := initTracing(t, "micro-server")
	if err != nil {
		t.Fatal(err)
	}

	if closer != nil {
		defer closer.Close()
	}
	runMicroServer()
}

type Greeter struct{}

func (g *Greeter) Hello(ctx context.Context, req *proto.Request, rsp *proto.Response) error {
	formatString(ctx, "evan")
	rsp.Greeting = "Hello " + req.Name
	return nil
}

func formatString(ctx context.Context, helloTo string) {
	span, _ := opentracing.StartSpanFromContext(ctx, "formatString")
	defer span.Finish()

	helloStr := fmt.Sprintf("hello, %s", helloTo)
	span.LogFields(
		tracinglog.String("event", "string-format"),
		tracinglog.String("value", helloStr),
	)
}

// Setup and the client
func runClient(cli client.Client) {
	// Create new greeter client
	greeter := proto.NewGreeterService("greeter", cli)

	// Call the greeter
	rsp, err := greeter.Hello(context.TODO(), &proto.Request{Name: "John"})
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print response
	fmt.Println(rsp.Greeting)
}

func runMicroServer() {
	service := grpcsvc.NewService(
		microsvc.Name("greeter"),
		microsvc.Registry(memory.NewRegistry()),
		microsvc.WrapHandler(NewHandlerWrapper(opentracing.GlobalTracer())),
		microsvc.WrapClient(NewClientWrapper(opentracing.GlobalTracer())),
	)

	service.Init()

	// Register handler
	proto.RegisterGreeterHandler(service.Server(), new(Greeter))

	go func() {
		// Run the server
		if err := service.Run(); err != nil {
			fmt.Println(err)
		}
	}()

	time.Sleep(time.Second * 3)
	runClient(service.Client())
}
