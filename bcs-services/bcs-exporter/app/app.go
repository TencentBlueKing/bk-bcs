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
	"errors"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/common/metric"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-exporter/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-exporter/pkg/dbus"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-exporter/pkg/rdiscover"
	"io/ioutil"
	"net/http"
	"strconv"

	restful "github.com/emicklei/go-restful"
)

// BcsExporter BcsExporter管理器定义
type BcsExporter struct {
	conf        *config.Config
	bus         dbus.MsgBusIf
	acts        []*httpserver.Action
	rddiscovery *rdiscover.RDiscover
	httpServ    *httpserver.HttpServer
}

// RegisterAction  register http request action
func (cli *BcsExporter) RegisterAction(method string, action *httpserver.Action) error {

	cli.acts = append(cli.acts, action)
	return nil
}

// Do do something
func (cli *BcsExporter) Do(req *restful.Request, resp *restful.Response) {

	dataid, dataidErr := strconv.Atoi(req.PathParameter("dataid"))
	if nil != dataidErr {
		blog.Error("failed to convert the dataid")
		resp.WriteError(http.StatusBadRequest, dataidErr)
		return
	}

	typeid, typeidErr := strconv.Atoi(req.PathParameter("type"))
	if nil != typeidErr {
		blog.Error("failed to convert the type")
		resp.WriteError(http.StatusBadRequest, typeidErr)
		return
	}

	rpy, err := ioutil.ReadAll(req.Request.Body)
	// TODO: need close body
	if err != nil {
		blog.Error("failed to read body , error info is %s", err.Error())
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	_, err = cli.bus.Write(typeid, dataid, rpy)
	if nil != err {
		blog.Error("failed to write data , error info is %s", err.Error())
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteError(http.StatusOK, errors.New("success"))
}

// Run 执行启动逻辑
func (cli *BcsExporter) Run() error {

	if err := cli.rddiscovery.Start(); nil != err {
		blog.Errorf("failed to start rdiscovery, error is %s", err.Error())
		return err
	}

	// actions register
	cli.RegisterAction("POST", httpserver.NewAction("POST", "/export/type/{type}/data/{dataid}", nil, cli.Do))

	//http server
	chErr := make(chan error, 1)

	cli.httpServ.RegisterWebServer("/api/{apiversion}", nil, cli.acts)

	go func() {
		err := cli.httpServ.ListenAndServe()
		blog.Error("http listen and service failed! err:%s", err.Error())
		chErr <- err
	}()

	metricHandler(cli.conf)

	select {
	case err := <-chErr:
		blog.Error("exit!, err: %s", err.Error())
		return err
	}
}

// newExporter 创建服务对象
func newExporter(cfg *config.Config) (*BcsExporter, error) {

	bcsExporter := &BcsExporter{}
	bcsExporter.conf = cfg
	bus, busErr := dbus.New(cfg)
	if nil != busErr {
		return nil, busErr
	}

	bcsExporter.bus = bus

	bcsExporter.httpServ = httpserver.NewHttpServer(cfg.ListenPort, cfg.ListenIP, "")

	if cfg.ServCertDir != "" {
		if cfg.ServCert.CertFile == "" || cfg.ServCert.KeyFile == "" {
			cfg.ServCert.CertFile = cfg.ServCertDir + "/bcs-inner-server.crt"
			cfg.ServCert.KeyFile = cfg.ServCertDir + "/bcs-inner-server.key"
			cfg.ServCert.CAFile = cfg.ServCertDir + "/bcs-inner-ca.crt"
			cfg.ServCert.IsSSL = true
		}
	}

	if cfg.ServCert.IsSSL {
		bcsExporter.httpServ.SetSsl(cfg.ServCert.CAFile, cfg.ServCert.CertFile, cfg.ServCert.KeyFile, cfg.ServCert.CertPasswd)
	}

	bcsExporter.rddiscovery = rdiscover.NewRDiscover(cfg.ZKServerAddress, cfg.ListenIP, cfg.ListenPort, cfg.MetricPort, cfg.ServCert.IsSSL)
	return bcsExporter, nil
}

// Run 实例化进程配置
func Run(cfg *config.Config) error {

	if err := dbus.RegisterOutputPlugins(cfg); err != nil {
		return err
	}

	bcsExporter, err := newExporter(cfg)
	if nil != err {
		return err
	}
	if err := bcsExporter.Run(); nil != err {
		return err
	}

	return fmt.Errorf("can not run here")
}

func metricHandler(op *config.Config) {
	c := metric.Config{
		ModuleName: types.BCS_MODULE_EXPORTER,
		MetricPort: op.MetricPort,
		IP:         op.ListenIP,
		RunMode:    metric.Master_Master_Mode,

		SvrCaFile:   op.ServCert.CAFile,
		SvrCertFile: op.ServCert.CertFile,
		SvrKeyFile:  op.ServCert.KeyFile,
		SvrKeyPwd:   op.ServCert.CertPasswd,
	}

	if err := metric.NewMetricController(
		c,
		func() metric.HealthMeta {
			return metric.HealthMeta{
				IsHealthy:   true,
				Message:     "",
				CurrentRole: metric.SlaveRole,
			}
		},
	); err != nil {
		blog.Errorf("metric server error: %v", err)
		return
	}
	blog.Infof("start metric server successfully")
}
