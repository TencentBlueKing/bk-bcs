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

package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/ssl"
	"bk-bcs/bcs-common/common/zkclient"
	"bk-bcs/bcs-services/bcs-health/master/app/config"
	"bk-bcs/bcs-services/bcs-health/pkg/job"
	"bk-bcs/bcs-services/bcs-health/pkg/job/processor"

	"bk-bcs/bcs-common/common/static"
	"bk-bcs/bcs-services/bcs-health/pkg/alarm/utils"
	"bk-bcs/bcs-services/bcs-health/pkg/healthz"
	"bk-bcs/bcs-services/bcs-health/pkg/role"
	etcdc "github.com/coreos/etcd/client"
	"github.com/emicklei/go-restful"
)

func NewHttpAlarm(c config.Config, alarm utils.AlarmFactory, etcdCli etcdc.KeysAPI, role role.RoleInterface) (*HttpAlarm, error) {
	addrs := strings.Split(c.BCSZk, ",")
	jobCtrl, err := job.NewJobController(addrs)
	if err != nil {
		return nil, fmt.Errorf("new job controller failed. err: %v", err)
	}

	cli, err := healthz.NewHealthzClient(addrs, c.CertConfig)
	if err != nil {
		return nil, fmt.Errorf("new healthz client failed, err %v", err)
	}

	healthCtl, err := healthz.NewHealthCtrl(cli)
	if err != nil {
		return nil, fmt.Errorf("new healthzctl failed. err: %v", err)
	}

	processor, err := processor.NewJobProcessor(c.LocalIP, c.ETCD.EtcdRootPath, etcdCli, alarm, role)
	if err != nil {
		return nil, fmt.Errorf("new job processor failed, err: %v", err)
	}
	httpAlarm := HttpAlarm{
		s: &Server{
			sendTo:       alarm,
			zkClient:     zkclient.NewZkClient(strings.Split(c.BCSZk, ",")),
			jobCtrl:      jobCtrl,
			jobProcessor: processor,
			healthzCtrl:  healthCtl,
		},
	}

	api := new(restful.WebService).Path("/bcshealth/v1").Produces(restful.MIME_JSON)
	container := restful.NewContainer()
	container.Add(api)
	api.Route(api.POST("create").To(httpAlarm.Create))
	api.Route(api.POST("sendalarm").To(httpAlarm.CreateAlarm))
	api.Route(api.POST("setmaintenance").To(httpAlarm.SetMaintenance))
	api.Route(api.POST("cancelmaintenance").To(httpAlarm.CancelMaintenance))
	api.Route(api.GET("watchjobs").To(httpAlarm.WatchJobs))
	api.Route(api.GET("listjobs").To(httpAlarm.ListJobs))
	api.Route(api.POST("reportjobs").To(httpAlarm.ReportJobs))
	api.Route(api.GET("healthz").To(httpAlarm.HealthZ))
	api.Route(api.GET("bcshealthz").To(httpAlarm.GetPlatformAndComponentHealthz))

	if len(c.ServerCertFile) == 0 &&
		len(c.ServerKeyFile) == 0 &&
		len(c.CAFile) == 0 {
		blog.Infof("start insecure serve on %s:%d", c.Address, c.Port)
		go func() {
			insecureServer := &http.Server{
				Addr:    net.JoinHostPort(c.Address, strconv.FormatUint(uint64(c.Port), 10)),
				Handler: container,
			}
			if err := insecureServer.ListenAndServe(); nil != err {
				blog.Fatal(err)
			}
		}()
		return &httpAlarm, nil
	}

	// user https
	ca, err := ioutil.ReadFile(c.CAFile)
	if nil != err {
		return nil, fmt.Errorf("read server tls file failed. err:%v", err)
	}
	capool := x509.NewCertPool()
	capool.AppendCertsFromPEM(ca)
	tlsconfig, err := ssl.ServerTslConfVerityClient(c.CAFile,
		c.ServerCertFile,
		c.ServerKeyFile,
		static.ServerCertPwd)
	if err != nil {
		return nil, fmt.Errorf("generate tls config failed. err: %v", err)
	}
	tlsconfig.BuildNameToCertificate()

	blog.Info("start secure serve on %s:%d", c.Address, c.Port)

	ln, err := net.Listen("tcp", net.JoinHostPort(c.Address, strconv.FormatUint(uint64(c.Port), 10)))
	if err != nil {
		return nil, fmt.Errorf("listen secure server failed. err: %v", err)
	}
	listener := tls.NewListener(ln, tlsconfig)
	go func() {
		if err := http.Serve(listener, container); nil != err {
			blog.Fatalf("server https failed. err: %v", err)
		}
	}()

	return &httpAlarm, nil
}

type Server struct {
	sendTo       utils.AlarmFactory
	zkClient     *zkclient.ZkClient
	jobCtrl      job.JobInterf
	jobProcessor processor.JobProcessor
	healthzCtrl  *healthz.HealthzCtrl
}

type HttpAlarm struct {
	s *Server
}

func (r *HttpAlarm) Run(stop <-chan struct{}) error {
	return nil
}

func (r *HttpAlarm) HealthZ(req *restful.Request, resp *restful.Response) {
	resp.WriteHeader(http.StatusOK)
	ok := `{"status":"ok"}`
	resp.Write([]byte(ok))
}
