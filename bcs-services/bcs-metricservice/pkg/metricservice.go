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

package pkg

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/common/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/health"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/rdiscover"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/route"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/zk"
)

// StorageServer is a data struct of bcs storage server
type MetricServer struct {
	conf       *config.Config
	httpServer *httpserver.HttpServer
	rd         *rdiscover.RDiscover
}

// NewStorageServer create storage server object
func NewMetricServer(op *config.Config) (*MetricServer, error) {
	s := &MetricServer{}

	// Configuration
	s.conf = op

	// Http server
	s.httpServer = httpserver.NewHttpServer(s.conf.Port, s.conf.Address, "")
	if s.conf.ServCert.IsSSL {
		s.httpServer.SetSsl(s.conf.ServCert.CAFile, s.conf.ServCert.CertFile, s.conf.ServCert.KeyFile, s.conf.ServCert.CertPasswd)
	}

	// RDiscover
	s.rd = rdiscover.NewRDiscover(s.conf)

	// ApiResource
	a := api.GetAPIResource()
	if err := a.SetConfig(op, s.rd); err != nil {
		return nil, err
	}
	a.InitActions()

	return s, nil
}

func (s *MetricServer) initHTTPServer() error {
	a := api.GetAPIResource()

	// Api v1
	s.httpServer.RegisterWebServer(api.PathV1, nil, a.ActionsV1)
	return nil
}

// Start to run storage server
func (s *MetricServer) Start() error {
	chErr := make(chan error, 1)

	s.initHTTPServer()
	go func() {
		err := s.httpServer.ListenAndServe()
		blog.Errorf("http listen and service failed! err:%s", err.Error())
		chErr <- err
	}()

	metricHandler(s.conf)

	// register discover and elect
	roleEvent, err := s.rd.Start()
	if err != nil {
		blog.Error("failed to start discovery service: %v", err)
		return err
	}

	storageMgr, err := storage.New(s.conf, s.rd)
	if err != nil {
		blog.Error("failed to get storage manager: %v", err)
		return err
	}
	routeMgr, err := route.New(s.conf, s.rd)
	if err != nil {
		blog.Error("failed to get route manager: %v", err)
		return err
	}
	zkMgr, err := zk.New(s.conf)
	if err != nil {
		blog.Error("failed to get zk manager: %v", err)
		return err
	}

	metricManager := manager.NewMetricManager(roleEvent, s.conf, storageMgr, routeMgr, zkMgr)
	metricManager.Run()
	return nil
}

func metricHandler(op *config.Config) {
	c := metric.Config{
		ModuleName: "bcs-metricservice",
		MetricPort: op.MetricPort,
		IP:         op.Address,
		RunMode:    metric.Master_Slave_Mode,

		SvrCaFile:   op.ServCert.CAFile,
		SvrCertFile: op.ServCert.CertFile,
		SvrKeyFile:  op.ServCert.KeyFile,
		SvrKeyPwd:   op.ServCert.CertPasswd,
	}

	if err := metric.NewMetricController(
		c,
		health.GetHealth,
	); err != nil {
		blog.Errorf("metric server error: %v", err)
		return
	}
	blog.Infof("start metric server successfully")
}
