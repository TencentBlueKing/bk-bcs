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

package register

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	regd "github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
)

func Register(localIP string, scheme string, port uint, metricPort uint, zkAddr string, clusterid string) error {
	disc := regd.NewRegDiscoverEx(zkAddr, time.Duration(5*time.Second))
	if err := disc.Start(); nil != err {
		return fmt.Errorf("start discover service failed. Error:%v", err)
	}
	hostname, err := os.Hostname()
	if nil != err {
		blog.Errorf("get hostname failed. err: %v", err)
	}

	info := types.BcsHealthInfo{
		ServerInfo: types.ServerInfo{
			IP:         localIP,
			Port:       port,
			MetricPort: metricPort,
			HostName:   hostname,
			Scheme:     scheme,
			Version:    version.GetVersion(),
			Cluster:    clusterid,
			Pid:        os.Getpid(),
		},
	}

	js, err := json.Marshal(info)
	if nil != err {
		return fmt.Errorf("marshal health info failed. err: %v", err)
	}

	path := fmt.Sprintf("%s/%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_METRICCOLLECTOR, localIP)
	if err := disc.RegisterAndWatchService(path, js); nil != err {
		return fmt.Errorf("register service failed. Error:%v", err)
	}
	return nil
}

type RequestInfo struct {
	Module string `json:"module"`
	IP     string `json:"ip"`
}

type RespInfo struct {
	RequestInfo
	ClusterID string `json:"clusterid"`
	Extension string `json:"extendedinfo"`
}

type Response struct {
	ErrorCode int      `json:"errcode"`
	ErrorMsg  string   `json:"errmsg"`
	Resp      RespInfo `json:"data"`
}
