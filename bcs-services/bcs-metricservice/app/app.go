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
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg"
)

// BcsMetricService BcsMetricService管理器定义
type BcsMetricService struct {
	acts     []*httpserver.Action
	httpServ *httpserver.HttpServer
}

// RegisterAction  register http request action
func (cli *BcsMetricService) RegisterAction(method string, action *httpserver.Action) error {

	cli.acts = append(cli.acts, action)
	return nil
}

// Run 执行启动逻辑
func (cli *BcsMetricService) Run() error {

	//http server
	chErr := make(chan error, 1)

	cli.httpServ.RegisterWebServer("/api/{apiversion}", nil, cli.acts)

	go func() {
		err := cli.httpServ.ListenAndServe()
		blog.Error("http listen and service failed! err:%s", err.Error())
		chErr <- err
	}()

	select {
	case err := <-chErr:
		blog.Error("exit!, err: %s", err.Error())
		return err
	}

}

// Run 实例化进程配置
func Run(cfg *config.Config) error {

	server, err := pkg.NewMetricServer(cfg)
	if err != nil {
		blog.Error("fail to create storage server. err:%s", err.Error())
		return err
	}

	if err := common.SavePid(cfg.ProcessConfig); err != nil {
		blog.Warn("fail to save pid. err:%s", err.Error())
	}

	return server.Start()
}
