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

package httpsvr

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

func (h *HttpServerClient) readinessProbe(request *restful.Request, response *restful.Response) {
	startTime := time.Now()
	mf := func(status string) {
		metrics.ReportAPIRequestMetric("readiness_probe", "GET", status, startTime)
	}

	if err := h.NodePortBindCache.Init(); err != nil {
		blog.Warnf("init nodeportbinding cache failed, err: %s", err.Error())
		_ = response.WriteError(http.StatusInternalServerError, fmt.Errorf("list port pool failed, err: %s",
			err.Error()))
		mf(strconv.Itoa(http.StatusInternalServerError))
		return
	}

	mf(strconv.Itoa(http.StatusOK))
	data := CreateResponseData(nil, "success", nil)
	_, _ = response.Write(data)
}
