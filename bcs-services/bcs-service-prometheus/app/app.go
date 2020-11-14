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

package app

import (
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-service-prometheus/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-service-prometheus/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-service-prometheus/controller"
)

//Run the prometheus controller
func Run(op *options.PrometheusControllerOption) error {

	conf := &config.Config{}
	setConfig(conf, op)
	//pid
	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}

	server := controller.NewPrometheusController(conf)
	err := server.Start()
	if err != nil {
		return err
	}

	blog.Info("app start PrometheusController server ... ")
	return nil
}

func setConfig(conf *config.Config, op *options.PrometheusControllerOption) {
	conf.ClusterZk = op.ClusterZk
	conf.PromFilePrefix = op.PromFilePrefix
	conf.ClusterID = op.ClusterID
	conf.CadvisorPort = op.CadvisorPort
	conf.NodeExportPort = op.NodeExporterPort
	conf.ServiceZk = op.BCSZk
	conf.EnableMesos = op.EnableMesos
	conf.EnableNode = op.EnableNode
	conf.EnableService = op.EnableService
	conf.Kubeconfig = op.Kubeconfig
	conf.ServiceModules = op.ServiceModules
	conf.ClusterModules = op.ClusterModules
	conf.EnableServiceMonitor = op.EnableServiceMonitor
}
