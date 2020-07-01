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

package driver

import (
	"fmt"
	btypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/route"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/zk"

	simplejson "github.com/bitly/go-simplejson"
)

type ClusterDriver interface {
	GetCollectorTypeName() string
	GetIPMeta() (map[string]btypes.ObjectMeta, error)
	GetApplicationJson(imageBase string) (*simplejson.Json, error)
	CreateApplication(data []byte) error
	DeleteApplication(data []byte) error
}

func GetClusterDriver(m *types.Metric, c *config.Config, s storage.Storage, r route.Route, z zk.Zk) (ClusterDriver, error) {
	ct := types.GetClusterType(m.ClusterType)
	switch ct {
	case types.ClusterMesos:
		return NewMesosDriver(m, c, s, r), nil
	case types.ClusterK8S:
		return NewK8SDriver(m, c, s, r), nil
	case types.BcsComponents:
		return NewComponentsDriver(c, z), nil
	default:
		return nil, fmt.Errorf("get cluster driver failed, unknown cluster type: %s", m.ClusterType)
	}
}
