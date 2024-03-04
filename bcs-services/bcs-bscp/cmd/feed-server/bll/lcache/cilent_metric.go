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

package lcache

import (
	clientset "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/client-set"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/cache-service"
)

// newClientMetric xxx
func newClientMetric(mc *metric, cs *clientset.ClientSet) *ClientMetric {
	cm := new(ClientMetric)
	cm.mc = mc
	cm.cs = cs
	return cm
}

// ClientMetric xxx
type ClientMetric struct {
	mc *metric
	cs *clientset.ClientSet
}

// Set tore client metric data into redis queues
func (cm *ClientMetric) Set(kt *kit.Kit, bizID, appID uint32, payload []byte) error {
	_, err := cm.cs.CS().SetClientMetric(kt.Ctx, &pbcs.SetClientMetricReq{
		BizId:   bizID,
		AppId:   appID,
		Payload: payload,
	})
	if err != nil {
		return err
	}
	return nil
}
