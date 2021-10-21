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

package processor

import (
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/filter"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/server"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/server/proxier"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/server/resthdrs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/processor/http/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/processor/http/actions/v4http/mesos/webconsole"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/tunnel"
)

type Processor struct {
	config   *config.ApiServConfig
	httpServ *httpserver.HttpServer
}

func NewProcessor(conf *config.ApiServConfig) *Processor {
	proc := &Processor{
		config:   conf,
		httpServ: httpserver.NewHttpServer(conf.Port, conf.Address, conf.Sock),
	}

	if conf.ServCert.IsSSL {
		proc.httpServ.SetSsl(conf.ServCert.CAFile, conf.ServCert.CertFile, conf.ServCert.KeyFile, conf.ServCert.CertPasswd)
	}

	proc.httpServ.SetInsecureServer(conf.InsecureAddress, conf.InsecurePort)

	return proc
}

func (p *Processor) Start() error {
	server.Setup(p.config)
	server.StartRbacSync(p.config)

	//handler http service
	generalFilter, err := filter.NewFilter(p.config)
	if err != nil {
		blog.Errorf("new filter failed: %v", err)
		os.Exit(1)
	}

	proxier.DefaultReverseProxyDispatcher.Initialize()

	tunnelServer := tunnel.NewTunnelServer()
	err = tunnel.StartPeerManager(p.config, tunnelServer)
	if err != nil {
		blog.Errorf("failed to start peermanager: %s", err.Error())
		return err
	}

	p.httpServ.RegisterWebServer("", generalFilter.Filter, actions.GetApiAction())
	router := p.httpServ.GetRouter()
	webContainer := p.httpServ.GetWebContainer()
	router.Handle("/bcsapi/v1/websocket/connect", tunnelServer)
	router.Handle("/bcsapi/v1/webconsole/{sub_path:.*}", webconsole.NewWebconsoleProxy(p.config.ClientCert))
	//mesos clueter api forwarding
	router.Handle("/bcsapi/{sub_path:.*}", webContainer)
	//kubernetes cluster api forwarding
	router.Handle("/rest/{sub_path:.*}", resthdrs.CreateRestContainer("/rest"))
	router.Handle("/tunnels/clusters/{cluster_identifier}/{sub_path:.*}", proxier.DefaultReverseProxyDispatcher)
	if err := p.httpServ.ListenAndServeMux(p.config.VerifyClientTLS); err != nil {
		blog.Errorf("http ListenAndServe error %s", err.Error())
		os.Exit(1)
	}

	return nil
}
