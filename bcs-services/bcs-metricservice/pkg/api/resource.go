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

package api

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/rdiscover"
)

type MetricAPIResource struct {
	Conf      *config.Config
	Rd        *rdiscover.RDiscover
	ActionsV1 []*httpserver.Action
}

var api = MetricAPIResource{}

func GetAPIResource() *MetricAPIResource {
	return &api
}

func (a *MetricAPIResource) InitActions() {
	a.ActionsV1 = append(a.ActionsV1, GetApiV1Action()...)
}

func (a *MetricAPIResource) SetConfig(op *config.Config, rd *rdiscover.RDiscover) error {
	a.Conf = op
	a.Rd = rd
	return InitMetric()
}
