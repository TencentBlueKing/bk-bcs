/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

// Register must only be called after the metric service is started.
func Register() prometheus.Registerer {
	return globalRegister
}

// RegisterHTTPMetrics register http metrics to global reg
func RegisterHTTPMetrics() {
	globalRegister.MustRegister(httpRequestsTotal)
	globalRegister.MustRegister(httpRequestDuration)
}

// httpHandler used to expose the metrics to prometheus.
var httpHandler http.Handler

// Handler returns the http handler with metrics.
func Handler() http.Handler {
	return httpHandler
}

// collectHTTPRequestMetric http metrics 处理
func collectHTTPRequestMetric(handler, method, code string, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(handler, method, code).Inc()
	httpRequestDuration.WithLabelValues(handler, method, code).Observe(duration.Seconds())
}

// RequestCollect collect metrics by name middleware
func RequestCollect(name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()
			defer func() {
				collectHTTPRequestMetric(name, r.Method, strconv.Itoa(ww.Status()), time.Since(t1))
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
