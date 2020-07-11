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

package config

import (
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/util"
)

type Config struct {
	conf.FileConfig
	conf.ZkConfig
	conf.LocalConfig
	conf.CertConfig
	conf.MetricConfig
	conf.LicenseServerConfig
	conf.ProcessConfig
	conf.LogConfig
	// name of this health slave cluster. unique in all the health clusters.
	ClusterName string      `json:"cluster_name" value:"" usage:"name of this health cluster, must be unique among all the clusters."`
	Zones       []util.Zone `json:"zones" value:"" usage:"zone that this slave have. default is all zones."`
}

func ParseConfig() (Config, error) {
	c := new(Config)
	conf.Parse(c)
	if len(c.LocalIP) == 0 || len(c.BCSZk) == 0 ||
		len(c.ClusterName) == 0 {
		return *c, errors.New("invalid configuration")
	}

	if len(c.Zones) == 0 {
		fmt.Printf("do not config zones, use all zones as default.\n")
		c.Zones = util.Zones{util.AllZones}
	}
	return *c, nil
}
