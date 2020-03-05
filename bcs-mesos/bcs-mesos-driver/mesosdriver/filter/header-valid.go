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

package filter

import (
	"fmt"

	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-mesos/bcs-mesos-driver/mesosdriver/config"

	"github.com/emicklei/go-restful"
)

type HeaderValidFilter struct {
	conf *config.MesosDriverConfig
}

func NewHeaderValidFilter(conf *config.MesosDriverConfig) RequestFilterFunction {
	return &HeaderValidFilter{
		conf: conf,
	}
}

func (h *HeaderValidFilter) Execute(req *restful.Request) (int, error) {
	clusterId := req.Request.Header.Get("BCS-ClusterID")
	if clusterId != h.conf.Cluster {
		return common.BcsErrMesosDriverHttpFilterFailed, fmt.Errorf("http header BCS-ClusterID %s don't exist", clusterId)
	}

	return 0, nil
}
