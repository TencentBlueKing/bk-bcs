/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package app

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/http/pprof"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pbcloudnetservice "github.com/Tencent/bk-bcs/bcs-network/api/protocol/cloudnetservice"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/cleaner"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/cloud/aws"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/cloud/tencentcloud"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/metric"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/option"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/store/kube"
	"github.com/Tencent/bk-bcs/bcs-network/pkg/leaderelection"
)

// CloudNetservice object for bcs cloud netservice
type CloudNetservice struct {
	// config for app cloud netservice
	cfg *option.Config

	// listener
	lis net.Listener

	// grpc server
	grpcServer *grpc.Server

	// metric endpoint
	metricEndpoint string

	// metric server
	metricServer *http.Server

	// metric collector
	metricCollector *metric.Collector

	// store interface
	storeIf store.Interface

	// cloud interface
	cloudIf cloud.Interface

	// ip cleaner
	ipCleaner *cleaner.IPCleaner

	// elector for leader election
	elector *leaderelection.Client

	// http mux
	mux *http.ServeMux
}

// NewCloudNetservice create cloud netservice app
func NewCloudNetservice(cfg *option.Config) *CloudNetservice {
	return &CloudNetservice{
		cfg: cfg,
	}
}

func (cn *CloudNetservice) initStore() error {
	blog.Infof("init store")
	kubeClient, err := kube.NewClient(cn.cfg.Kubeconfig)
	if err != nil {
		blog.Errorf("init store failed, err %s", cn.cfg.Kubeconfig)
		return err
	}
	cn.storeIf = kubeClient
	return nil
}

func (cn *CloudNetservice) initCloud() error {
	blog.Infof("init cloud api")
	var cloudIf cloud.Interface
	var err error
	switch cn.cfg.CloudMode {
	case cloud.CLOUD_TENCENT:
		cloudIf, err = tencentcloud.NewClient()
		if err != nil {
			blog.Errorf("create tencent cloud client failed, err %s", err.Error())
			return fmt.Errorf("create tencent cloud client failed, err %s", err.Error())
		}
	case cloud.CLOUD_AWS:
		cloudIf, err = aws.NewClient()
		if err != nil {
			blog.Errorf("create aws cloud client failed, err %s", err.Error())
			return fmt.Errorf("create aws cloud client failed, err %s", err.Error())
		}
	default:
		blog.Errorf("invalid cloud mode %s", cn.cfg.CloudMode)
		return fmt.Errorf("invalid cloud mode %s", cn.cfg.CloudMode)
	}
	cn.cloudIf = cloudIf
	return nil
}

func (cn *CloudNetservice) initLeaderElection() error {
	elector, err := leaderelection.New(resourcelock.LeasesResourceLock, 
		"bcs-cloud-netservice", "bcs-system", cn.cfg.Kubeconfig, 15*time.Second, 10*time.Second, 2*time.Second)
	if err != nil {
		return err
	}
	go elector.RunOrDie()

	cn.elector = elector
	return nil
}

func (cn *CloudNetservice) initSwagger() {
	if len(cn.cfg.SwaggerDir) != 0 {
		cn.mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, path.Join(cn.cfg.SwaggerDir, strings.TrimPrefix(r.URL.Path, "/swagger/")))
		})
	}
}

func (cn *CloudNetservice) initHTTPGateway() {
	gwmux := runtime.NewServeMux()

	err := pbcloudnetservice.RegisterCloudNetserviceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		cn.cfg.Address+":"+strconv.Itoa(int(cn.cfg.Port)),
		[]grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		blog.Fatalf("register cloud netservice gateway, err %s", err.Error())
	}

	// handle gateway.
	cn.mux.Handle("/", gwmux)
}

func (cn *CloudNetservice) initPProf() {
	if cn.cfg.Debug {
		blog.Infof("init pprof handler")
		cn.mux.HandleFunc("/debug/pprof/", pprof.Index)
		cn.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		cn.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		cn.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		cn.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
}

func (cn *CloudNetservice) initMetrics() {
	blog.Infof("init metrics handler")
	cn.metricEndpoint = cn.cfg.Address + ":" + strconv.Itoa(int(cn.cfg.MetricPort))
	cn.metricCollector = metric.NewCollector(cn.metricEndpoint, "/metrics")
	cn.metricCollector.Init()
	cn.metricCollector.RegisterMux(cn.mux)
}

func (cn *CloudNetservice) initIPCleaner() {
	blog.Infof("init ip cleaner")
	cn.ipCleaner = cleaner.NewIPCleaner(
		time.Duration(cn.cfg.IPMaxIdleMinute)*time.Minute, time.Duration(cn.cfg.IPCleanIntervalMinute)*time.Minute,
		cn.storeIf, cn.cloudIf, cn.elector)
	go cn.ipCleaner.Run(context.TODO())
}

func (cn *CloudNetservice) initModules() {

	if err := cn.initStore(); err != nil {
		blog.Fatalf("initStore failed, err %s", err.Error())
	}
	if err := cn.initCloud(); err != nil {
		blog.Fatalf("initCloud failed, err %s", err.Error())
	}
	if err := cn.initLeaderElection(); err != nil {
		blog.Fatalf("initLeaderElection failed, err %s", err.Error())
	}

	cn.initIPCleaner()

	cn.mux = http.NewServeMux()

	// init http gateway
	cn.initHTTPGateway()
	cn.initMetrics()
	cn.initPProf()
	cn.initSwagger()

	cn.metricServer = &http.Server{
		Addr:    cn.metricEndpoint,
		Handler: cn.mux,
	}

	go func() {
		blog.Infof("start metrics and pprof server")
		err := cn.metricServer.ListenAndServe()
		if err != nil {
			blog.Errorf("metric server Listen failed, err %s", err.Error())
		}
	}()
}

// Run run the server
func (cn *CloudNetservice) Run() {

	cn.initModules()

	lis, err := net.Listen("tcp",
		cn.cfg.Address+":"+strconv.Itoa(int(cn.cfg.Port)))
	if err != nil {
		blog.Fatalf("listen on endpoint failed, err %s", err.Error())
	}

	// run grpc server
	cn.grpcServer = grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32))
	pbcloudnetservice.RegisterCloudNetserviceServer(cn.grpcServer, cn)
	blog.Infof("registered cloud netservice grpc server")

	if err := cn.grpcServer.Serve(lis); err != nil {
		blog.Errorf("start grpc server with net listener failed, err %s", err.Error())
	}
}

// Stop stop the server
func (cn *CloudNetservice) Stop() {
	cn.grpcServer.GracefulStop()
	err := cn.metricServer.Shutdown(context.Background())
	if err != nil {
		blog.Errorf("shut down metric server failed, err %s", err.Error())
	}
}
