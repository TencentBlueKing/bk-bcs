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

package collector

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
)

const (
	TraditionalClusterID string = "bcs_unique_const_clusterid"
	TraditionalNamespace string = "bcs_unique_const_namespace"
	TraditionalName      string = "bcs_unique_const_name"
)

type rspRst struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Version string                 `json:"version"`
		Cfgs    [][]types.CollectorCfg `json:"cfgs"`
	} `json:"data"`
}

// Collector the collector configuration manager
type Collector interface {
	Run(ctx context.Context, cfg *config.Config) error
}
