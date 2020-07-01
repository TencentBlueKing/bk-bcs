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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/zk"

	simplejson "github.com/bitly/go-simplejson"
)

type ComponentsDriver struct {
	config *config.Config
	zk     zk.Zk
}

func NewComponentsDriver(config *config.Config, zk zk.Zk) ClusterDriver {
	return &ComponentsDriver{config: config, zk: zk}
}

func (cd *ComponentsDriver) GetCollectorTypeName() string {
	return ""
}

func (cd *ComponentsDriver) GetApplicationJson(string) (*simplejson.Json, error) {
	return nil, nil
}

func (cd *ComponentsDriver) CreateApplication(data []byte) error {
	return nil
}

func (cd *ComponentsDriver) DeleteApplication(data []byte) error {
	return nil
}

func (cd *ComponentsDriver) GetIPMeta() (map[string]btypes.ObjectMeta, error) {
	endpoints := cd.zk.List(cd.config.EndpointWatchPath)
	ipMeta := make(map[string]btypes.ObjectMeta)

	for _, endpoint := range endpoints {
		key := fmt.Sprintf("%s%s%d", endpoint.IP, IPPortGap, endpoint.Port)
		ipMeta[key] = btypes.ObjectMeta{
			Name:        endpoint.Name,
			NameSpace:   endpoint.Path,
			Annotations: map[string]string{types.BcsComponentsSchemeKey: endpoint.Scheme}}
	}
	return ipMeta, nil
}
