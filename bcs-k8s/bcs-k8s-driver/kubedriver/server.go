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

package kubedriver

import (
	"crypto/tls"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/metric"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-driver/kubedriver/custom"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-driver/kubedriver/options"
	disreg "github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-watch/pkg/discovery/register"
	"net"
	"net/http"

	restful "github.com/emicklei/go-restful"
)

const (
	DefaultKubeURLPrefix = "k8sdriver/v4"
	MetricPort           = 9090
	ModuleName           = "k8s-driver"
)

type DriverServer struct {
	RootWebContainer *restful.Container
	Options          *options.KubeDriverServerOptions
}

//NewDriverServer Create a new DriverServer instance
func NewDriverServer(o *options.KubeDriverServerOptions) DriverServer {
	return DriverServer{
		RootWebContainer: restful.NewContainer(),
		Options:          o,
	}
}

//StartServer start driver server
func StartServer(o *options.KubeDriverServerOptions) error {
	clusterID, err := GetClusterID(o)
	if err != nil {
		return err
	}

	server := NewDriverServer(o)

	if err := o.Validate(); err != nil {
		return err
	}

	processConf := conf.ProcessConfig{
		PidDir: "pid",
	}
	if err := common.SavePid(processConf); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
		return err
	}

	if o.Environment == "prod" || o.Environment == "stag" {

		if o.RegisterWithWebsocket {
			err := buildWebsocketToApi(o)
			if err != nil {
				blog.Fatalf("err when register with websocket: %s", err.Error())
				return err
			}
		} else {
			// Register node to zk and then keep it registered
			// Get current node info to register it to zookeeper
			serverInfo := GetServerInfo(o, clusterID)
			node := custom.NewServiceNode(serverInfo)

			// Register current node info to zookeeper and start discovering
			basePath := fmt.Sprintf("%s/%s/%s",
				types.BCS_SERV_BASEPATH,
				types.BCS_MODULE_KUBERNETEDRIVER,
				clusterID,
			)
			reg := disreg.NewNodeRegister(o.ZkServers, basePath, &node)
			if err := reg.DoRegister(); err != nil {
				return fmt.Errorf("unable to register driver: %s", err)
			}
			go reg.StartDiscover(0)
		}
	}

	proxier := NewKubeSmartProxier(o.KubeMasterUrl, o.KubeClientTLS)
	version, err := proxier.RequestServerVersion()
	if err != nil {
		blog.Fatalf("can not get api server version: %s", err.Error())
		return err
	}
	blog.Infof("Kube API Server version: %s", version)

	// fetch api prefer
	proxier.RequestAPIPrefer()

	// Register webservice and routes
	ws := new(restful.WebService)
	ws.Produces(restful.MIME_JSON)

	proxier.RegisterToWS(ws)

	// add healthz
	ws.Route(ws.GET("/healthz").To(healthz))
	server.RootWebContainer.Add(ws)

	if o.SecureServerConfigured() {
		ServerTLSConfig, err := o.ServerTLS.ToConfigObj()
		if err != nil {
			return fmt.Errorf("unable to load server TLS Config: %s", err)
		}

		blog.Infof("Serving secure serve on %s", o.MakeServerAddress(options.ServerTypeSecure))
		ln, err := net.Listen("tcp", o.MakeServerAddress(options.ServerTypeSecure))
		if err != nil {
			return fmt.Errorf("listen secure server failed. err: %v", err)
		}

		listener := tls.NewListener(ln, ServerTLSConfig)
		// Start a go routine to serve in background
		go func() {
			if err := http.Serve(listener, server.RootWebContainer); nil != err {
				blog.Fatalf("server https failed. err: %v", err)
			}
		}()
	}

	if o.InsecureServerConfigured() {
		insecureServer := &http.Server{
			Addr:    o.MakeServerAddress(options.ServerTypeInsecure),
			Handler: server.RootWebContainer,
		}
		blog.Infof("Serving insecure server on %s", o.MakeServerAddress(options.ServerTypeInsecure))
		// Start a go routine to serve in background
		go func() {
			if err := insecureServer.ListenAndServe(); nil != err {
				blog.Fatalf("server http failed. err: %v", err)
			}
		}()
	}
	// add metric
	blog.Infof("start metric...")
	go startMetric(server.Options.HostIP, clusterID)

	select {}
}

func healthz(request *restful.Request, response *restful.Response) {
	custom.CustomSuccessResponse(response, "I AM ALIVE AND HAPPY", nil)
	return
}

func startMetric(moduleIP, clusterID string) error {
	// config the metric
	c := metric.Config{
		ModuleName:          ModuleName,
		IP:                  moduleIP,
		MetricPort:          MetricPort,
		DisableGolangMetric: true,
		ClusterID:           clusterID,
	}

	// check health
	healthz := func() metric.HealthMeta {
		return metric.HealthMeta{
			CurrentRole: "Master",
			IsHealthy:   true,
		}
	}
	if err := metric.NewMetricController(c, healthz); err != nil {
		blog.Fatalf("new metric collector failed. err: %v\n", err)
		return err
	}
	return nil
}
