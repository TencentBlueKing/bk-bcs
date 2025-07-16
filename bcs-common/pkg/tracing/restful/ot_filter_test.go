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

package restful

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	restful "github.com/emicklei/go-restful/v3"
	opentrace "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tracinglog "github.com/opentracing/opentracing-go/log"

	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/jaeger"
)

func initTracing(t *testing.T) (io.Closer, error) {
	tracer, err := tracing.NewInitTracing("server-restful",
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

func TestNewOTFilter(t *testing.T) {
	stopEveryThing := make(chan struct{}, 1)

	closer, err := initTracing(t)
	if err != nil {
		t.Fatal()
		return
	}

	if closer != nil {
		defer func(closer io.Closer) {
			_ = closer.Close()
		}(closer)
	}

	registerWebService()
	go func(ctx context.Context) {
		<-time.After(5 * time.Second)
		getServiceHello(ctx)
		stopEveryThing <- struct{}{}
	}(context.Background())

	<-stopEveryThing
	t.Log("server quit")
}

func registerWebService() {
	server := httpserver.NewHttpServer(8083, "0.0.0.0", "")

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("ListenAndServe failed: %v", err)
			return
		}
	}()

	filters := []restful.FilterFunction{}
	filters = append(filters, NewOTFilter(opentrace.GlobalTracer()), webserviceLogging)

	webService := server.NewWebService("/tracing", filters)
	webService.Route(webService.GET("/hello").To(hello))
}

// WebService Filter
func webserviceLogging(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	log.Printf("[webservice-filter (logger)] %s,%s\n", req.Request.Method, req.Request.URL)
	chain.ProcessFilter(req, resp)
	log.Println("webserviceLogging")
}

func hello(req *restful.Request, resp *restful.Response) {
	formatString(req.Request.Context(), "hello")
	resp.WriteHeader(http.StatusOK)
	_, _ = resp.Write([]byte("hello"))
}

func formatString(ctx context.Context, helloTo string) {
	span, _ := opentrace.StartSpanFromContext(ctx, "formatString")
	defer span.Finish()

	helloStr := fmt.Sprintf("hello, %s", helloTo)
	span.LogFields(
		tracinglog.String("event", "string-format"),
		tracinglog.String("value", helloStr),
	)
}

func getServiceHello(ctx context.Context) {
	span, _ := opentrace.StartSpanFromContext(ctx, "getServiceHello")

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1:8083/tracing/hello", nil)
	if err != nil {
		span.Finish()
		log.Fatal(err)
	}

	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, req.URL.Path)
	ext.HTTPMethod.Set(span, "GET")
	_ = span.Tracer().Inject(span.Context(), opentrace.HTTPHeaders, opentrace.HTTPHeadersCarrier(req.Header))

	_, err = client.Do(req)
	if err != nil {
		ext.LogError(span, err)
		span.Finish()
		log.Fatal(err)
	}
}
