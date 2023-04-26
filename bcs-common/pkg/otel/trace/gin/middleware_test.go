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

package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func init() {
	gin.SetMode(gin.ReleaseMode) // silence annoying log msgs
}

func TestChildSpanFromGlobalTracer(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr)))

	router := gin.New()
	router.Use(Middleware("test"))
	router.GET("/user/:id", func(c *gin.Context) {})

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)
	assert.Len(t, sr.Ended(), 1)
}

func TestWithFilter(t *testing.T) {
	t.Run("custom filter filtering route", func(t *testing.T) {
		sr := tracetest.NewSpanRecorder()
		otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr)))

		router := gin.New()
		f := func(req *http.Request) bool { return req.URL.Path != "/healthcheck" }
		router.Use(Middleware("test", WithFilter(f)))
		router.GET("/healthcheck", func(c *gin.Context) {})

		r := httptest.NewRequest("GET", "/healthcheck", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, r)
		assert.Len(t, sr.Ended(), 0)
	})

	t.Run("custom filter not filtering route", func(t *testing.T) {
		sr := tracetest.NewSpanRecorder()
		otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr)))

		router := gin.New()
		f := func(req *http.Request) bool { return req.URL.Path != "/healthcheck" }
		router.Use(Middleware("foobar", WithFilter(f)))
		router.GET("/user/:id", func(c *gin.Context) {})

		r := httptest.NewRequest("GET", "/user/123", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, r)
		assert.Len(t, sr.Ended(), 1)
	})
}
