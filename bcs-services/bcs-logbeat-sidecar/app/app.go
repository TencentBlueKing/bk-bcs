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
	"os"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-logbeat-sidecar/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-logbeat-sidecar/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-logbeat-sidecar/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-logbeat-sidecar/sidecar"
)

func Run(op *options.SidecarOption) error {

	conf := &config.Config{}
	setConfig(conf, op)
	//pid
	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}

	metric.NewMetricClient(op.Address, strconv.Itoa(int(op.MetricPort))).Start()
	blog.Info("app start Metric client on %s:%s", op.Address, strconv.Itoa(int(op.MetricPort)))

	controller, err := sidecar.NewSidecarController(conf)
	if err != nil {
		blog.Errorf("NewSidecarController error %s", err.Error())
		os.Exit(1)
	}

	controller.Start()
	blog.Info("app start SidecarController server ... ")
	return nil
}

func setConfig(conf *config.Config, op *options.SidecarOption) {
	conf.DockerSock = op.DockerSock
	conf.PrefixFile = op.PrefixFile
	conf.TemplateFile = op.TemplateFile
	conf.LogbeatDir = op.LogbeatDir
	conf.Kubeconfig = op.Kubeconfig
	conf.EvalSymlink = op.EvalSymlink
	conf.LogbeatPIDFilePath = op.LogbeatPIDFilePath
	conf.NeedReload = op.NeedReload
	conf.LogbeatOutputFormat = op.LogbeatOutputFormat
	if op.FileExtension == "" {
		conf.FileExtension = "yaml"
	} else {
		conf.FileExtension = op.FileExtension
	}
}
