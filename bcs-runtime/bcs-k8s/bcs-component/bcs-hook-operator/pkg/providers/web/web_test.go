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

package web

import (
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/util/testutil"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProvider_Run(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"result": {"name": "Tom","age": 18}}`))
	}))
	metric := hookv1alpha1.Metric{
		Name:             "test",
		SuccessCondition: "asInt(result) > 1",
		Provider: hookv1alpha1.MetricProvider{Web: &hookv1alpha1.WebMetric{
			URL:      ts.URL,
			JsonPath: "{$.result.age}",
		}},
	}
	c := NewWebMetricHttpClient(metric)
	j, err := NewWebMetricJsonParser(metric)
	if err != nil {
		t.Fatal(err)
	}
	p := NewWebMetricProvider(c, j)
	hr := testutil.NewHookRun("m0")
	mm := p.Run(hr, metric)
	assert.Equal(t, mm.Phase, hookv1alpha1.HookPhaseSuccessful)
}
