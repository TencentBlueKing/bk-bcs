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

package rdiscover

import (
	"bk-bcs/bcs-common/common/RegisterDiscover"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/common/version"

	"encoding/json"
	"os"
	"time"
)

//RDiscover route register and discover
type RDiscover struct {
	ip         string
	port       uint
	metricPort uint
	isSSL      bool
	rd         *RegisterDiscover.RegDiscover
}

//NewRDiscover create a object of RDiscover
func NewRDiscover(zkserv string, ip string, port, metricPort uint, isSSL bool) *RDiscover {
	return &RDiscover{
		ip:         ip,
		port:       port,
		metricPort: metricPort,
		isSSL:      isSSL,
		rd:         RegisterDiscover.NewRegDiscoverEx(zkserv, 10*time.Second),
	}
}

//Start the rdiscover
func (r *RDiscover) Start() error {

	//start regdiscover
	if err := r.rd.Start(); err != nil {
		blog.Error("fail to start register and discover serv. err:%s", err.Error())
		return err
	}

	//register self
	if err := r.registerSelf(); err != nil {
		blog.Error("fail to register route(%s). err:%s", r.ip, err.Error())
		return err
	}

	return nil
}

//Stop the rdiscover
func (r *RDiscover) Stop() error {

	r.rd.Stop()

	return nil
}

func (r *RDiscover) registerSelf() error {
	exporterServInfo := types.DataExporterInfo{}

	exporterServInfo.IP = r.ip
	exporterServInfo.Port = r.port
	exporterServInfo.MetricPort = r.metricPort
	exporterServInfo.Scheme = "http"
	if r.isSSL {
		exporterServInfo.Scheme = "https"
	}

	exporterServInfo.Version = version.GetVersion()
	exporterServInfo.Pid = os.Getpid()

	data, err := json.Marshal(exporterServInfo)
	if err != nil {
		blog.Error("fail to marshal exporterServInfo to json. err:%s", err.Error())
		return err
	}

	path := types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_EXPORTER + "/" + r.ip

	return r.rd.RegisterAndWatchService(path, data)
}
