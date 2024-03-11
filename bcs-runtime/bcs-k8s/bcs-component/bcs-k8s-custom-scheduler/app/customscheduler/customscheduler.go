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
 */

// Package customscheduler custom scheduler
package customscheduler

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-custom-scheduler/config"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-custom-scheduler/pkg/actions"
)

// CustomScheduler xxx
type CustomScheduler struct {
	config   *config.CustomSchedulerConfig
	httpServ *httpserver.HttpServer
}

// NewCustomScheduler creates an CustomScheduler object
func NewCustomScheduler(conf *config.CustomSchedulerConfig) *CustomScheduler {
	customSched := &CustomScheduler{
		config:   conf,
		httpServ: httpserver.NewHttpServer(conf.Port, conf.Address, conf.Sock),
	}

	if conf.ServCert.IsSSL {
		customSched.httpServ.SetSsl(
			conf.ServCert.CAFile,
			conf.ServCert.CertFile,
			conf.ServCert.KeyFile,
			conf.ServCert.CertPasswd)
	}

	customSched.httpServ.SetInsecureServer(conf.InsecureAddress, conf.InsecurePort)

	return customSched
}

// Start xxx
func (p *CustomScheduler) Start() error {

	p.httpServ.RegisterWebServer("", nil, actions.GetApiAction()) // nolint
	router := p.httpServ.GetRouter()
	webContainer := p.httpServ.GetWebContainer()
	router.Handle("/{sub_path:.*}", webContainer)
	if err := p.httpServ.ListenAndServeMux(p.config.VerifyClientTLS); err != nil {
		return fmt.Errorf("http ListenAndServe error %s", err.Error())
	}

	return nil
}
