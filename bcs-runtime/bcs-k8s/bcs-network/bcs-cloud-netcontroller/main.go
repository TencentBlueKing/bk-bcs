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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netcontroller/controllers"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netcontroller/internal/option"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netcontroller/pkg/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netcontroller/pkg/cloud/aws"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netcontroller/pkg/cloud/qcloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/grpclb"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"

	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = cloudv1.AddToScheme(scheme)
}

// initCloudNetservice init cloud netservice client
func initCloudNetservice(opts *option.ControllerOption) (pbcloudnet.CloudNetserviceClient, error) {
	conn, err := grpc.Dial(
		"",
		grpc.WithInsecure(),
		grpc.WithBalancer(grpc.RoundRobin(grpclb.NewPseudoResolver(opts.CloudNetServiceEndpoints))),
	)
	if err != nil {
		return nil, err
	}
	cloudNetClient := pbcloudnet.NewCloudNetserviceClient(conn)
	return cloudNetClient, nil
}

// initCloud init aws or tencent cloud client
func initCloud(opts *option.ControllerOption) (cloud.Interface, error) {
	var cloudClient cloud.Interface
	switch opts.Cloud {
	case constant.CloudKindTencent:
		cloudClient = qcloud.New()
	case constant.CloudKindAws:
		cloudClient = aws.New()
	default:
		return nil, fmt.Errorf("error cloud mode %s", opts.Cloud)
	}
	cloudClientWithMetric := cloud.NewCloudWithMetic(cloudClient)
	if err := cloudClientWithMetric.Init(); err != nil {
		return nil, fmt.Errorf("init cloud client failed, err %s", err.Error())
	}
	return cloudClientWithMetric, nil
}

func main() {
	opts := &option.ControllerOption{}

	var cloudNetserviceEndpoints string
	var verbosity int
	flag.StringVar(&opts.Address, "address", "127.0.0.1", "address for controller")
	flag.IntVar(&opts.MetricPort, "metric_port", 8081, "metric port for controller")
	flag.IntVar(&opts.Port, "port", 8080, "por for controller")
	flag.StringVar(&opts.Cloud, "cloud", "tencentcloud", "cloud mode for bcs network controller")
	flag.StringVar(&opts.Cluster, "cluster", "", "clusterid for bcs cluster")
	flag.StringVar(&cloudNetserviceEndpoints, "cloud_netservice_endpoints", "",
		"endpoints of cloud netservice, split by comma or semicolon")

	flag.StringVar(&opts.LogDir, "log_dir", "./logs", "If non-empty, write log files in this directory")
	flag.Uint64Var(&opts.LogMaxSize, "log_max_size", 500, "Max size (MB) per log file.")
	flag.IntVar(&opts.LogMaxNum, "log_max_num", 10, "Max num of log file.")
	flag.BoolVar(&opts.ToStdErr, "logtostderr", false, "log to standard error instead of files")
	flag.BoolVar(&opts.AlsoToStdErr, "alsologtostderr", false, "log to standard error as well as files")

	flag.IntVar(&verbosity, "v", 0, "log level for V logs")
	flag.StringVar(&opts.StdErrThreshold, "stderrthreshold", "2", "logs at or above this threshold go to stderr")
	flag.StringVar(&opts.VModule, "vmodule", "", "comma-separated list of pattern=N settings for file-filtered logging")
	flag.StringVar(&opts.TraceLocation, "log_backtrace_at", "", "when logging hits line file:N, emit a stack trace")

	flag.Parse()
	cloudNetserviceEndpoints = strings.Replace(cloudNetserviceEndpoints, ";", ",", -1)
	opts.CloudNetServiceEndpoints = strings.Split(cloudNetserviceEndpoints, ",")
	opts.Verbosity = int32(verbosity)

	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))

	blog.InitLogs(opts.LogConfig)
	defer blog.CloseLogs()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      opts.Address + ":" + strconv.Itoa(opts.MetricPort),
		LeaderElection:          true,
		LeaderElectionID:        "333fb49e.netconroller.bkbcs.tencent.com",
		LeaderElectionNamespace: "bcs-system",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	cloudClient, err := initCloud(opts)
	if err != nil {
		blog.Errorf("init cloud failed, err %s", err.Error())
		os.Exit(1)
	}

	cloudNetserviceClient, err := initCloudNetservice(opts)
	if err != nil {
		blog.Errorf("init cloud netservice failed, err %s", err.Error())
		os.Exit(1)
	}

	if err = (&controllers.NodeNetworkReconciler{
		Client:         mgr.GetClient(),
		CloudClient:    cloudClient,
		CloudNetClient: cloudNetserviceClient,
		Option:         opts,
		NodeEventer:    mgr.GetEventRecorderFor("bcs-cloud-netcontroller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create NodeNetwork controller", "controller", "NodeMetwork")
		os.Exit(1)
	}

	if err = (&controllers.NodeReconciler{
		Client:      mgr.GetClient(),
		CloudClient: cloudClient,
		Log:         ctrl.Log.WithName("controllers").WithName("Node"),
		Option:      opts,
		NodeEventer: mgr.GetEventRecorderFor("bcs-cloud-netcontroller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create k8s Node controller", "controller", "Node")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	if err = (&controllers.FixedIPReconciler{
		Ctx:    ctx,
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("FixedIP"),
		Option: opts,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "FixedIP")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}

	cancel()
}
