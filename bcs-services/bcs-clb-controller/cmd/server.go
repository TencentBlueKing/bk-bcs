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

package cmd

import (
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/conf"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/common"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/metric"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/processor"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/spf13/cobra"
)

var port int
var serviceRegistry string
var clbName string
var netType string
var backendIPType string
var kubeconfig string
var updateInterval int
var syncPeriod int

var logDir string
var logMaxSize uint64
var logMaxNum int
var toStdErr bool
var alsoToStdErr bool
var verbosity int32
var stdErrThreshold string
var vModule string
var traceLocation string

func init() {
	rootCmd.AddCommand(serverCmd)
	// option for clb controller server process
	serverCmd.Flags().IntVar(&port, "port", 18080, "port for clb controller server")
	serverCmd.Flags().StringVar(&serviceRegistry, "serviceRegistry", "kubernetes", "service registry for clb controller, available: [kubernetes, custom, mesos]")
	serverCmd.Flags().StringVar(&clbName, "clbname", "", "lb name for qcloud clb")
	serverCmd.Flags().StringVar(&netType, "netType", "private", "network type for clb, available: [private, public]")
	serverCmd.Flags().StringVar(&backendIPType, "backendIPType", "", "backend pod ip network type, available: [overlay, underlay]")

	serverCmd.Flags().IntVar(&updateInterval, "updateInterval", 5, "interval for update operations")
	// for kube registry
	serverCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig for access kube-apiserver, if empty, use in-cluster config, default: empty")
	serverCmd.Flags().IntVar(&syncPeriod, "syncPeriod", 60, "period for synchronize k8s services")

	// log option
	serverCmd.Flags().StringVar(&logDir, "log_dir", "./logs", "If non-empty, write log files in this directory")
	serverCmd.Flags().Uint64Var(&logMaxSize, "log_max_size", 500, "Max size (MB) per log file.")
	serverCmd.Flags().IntVar(&logMaxNum, "log_max_num", 10, "Max num of log file. The oldest will be removed if there is a extra file created.")
	serverCmd.Flags().BoolVar(&toStdErr, "logtostderr", false, "log to standard error instead of files")
	serverCmd.Flags().BoolVar(&alsoToStdErr, "alsologtostderr", false, "log to standard error as well as files")
	serverCmd.Flags().Int32Var(&verbosity, "v", 0, "log level for V logs")
	serverCmd.Flags().StringVar(&stdErrThreshold, "stderrthreshold", "2", "logs at or above this threshold go to stderr")
	serverCmd.Flags().StringVar(&vModule, "vmodule", "", "comma-separated list of pattern=N settings for file-filtered logging")
	serverCmd.Flags().StringVar(&traceLocation, "log_backtrace_at", "", "when logging hits line file:N, emit a stack trace")
}

func validateArgs() bool {
	if port < 0 || port > 65535 {
		blog.Infof("port %d invalid, must be in range [0, 65535]", port)
		return false
	}
	if serviceRegistry != "kubernetes" && serviceRegistry != "custom" && serviceRegistry != "mesos" {
		blog.Errorf("serviceRegistry %s invalid, must be in (kubernetes, custom, mesos)", serviceRegistry)
		return false
	}
	reg, _ := regexp.Compile("[a-zA-Z0-9-_\\.]+")
	if !reg.MatchString(clbName) {
		blog.Errorf("clbName %s invalid, must be [a-zA-Z0-9-\\.]+", clbName)
		return false
	}
	if netType != common.CLBNetTypePublic && netType != common.CLBNetTypePrivate {
		blog.Errorf("netType %s invalid, network type for clb, available: [private, public]", netType)
		return false
	}
	if backendIPType != common.BackendIPTypeOverlay && backendIPType != common.BackendIPTypeUnderlay {
		blog.Errorf("backendIPType %s invalid, backendIPType must be in (overlay, underlay)", backendIPType)
		return false
	}
	return true
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start clb controller server",
	Long: `Start clb controller server,
the server watch k8s services and clbIngresses to generate clb listener`,
	Run: func(cmd *cobra.Command, args []string) {

		if !validateArgs() {
			os.Exit(1)
		}

		blog.InitLogs(conf.LogConfig{
			LogDir:          logDir,
			LogMaxSize:      logMaxSize,
			LogMaxNum:       logMaxNum,
			ToStdErr:        toStdErr,
			AlsoToStdErr:    alsoToStdErr,
			Verbosity:       verbosity,
			StdErrThreshold: stdErrThreshold,
			VModule:         vModule,
			TraceLocation:   traceLocation,
		})
		blog.Infof("init log config done")
		clbCtrlOption := &processor.Option{
			Port:            port,
			ServiceRegistry: serviceRegistry,
			ClbName:         clbName,
			NetType:         netType,
			BackendIPType:   backendIPType,
			Kubeconfig:      kubeconfig,
			UpdatePeriod:    updateInterval,
			SyncPeriod:      syncPeriod,
		}

		//TODO: to create server with metric

		blog.Infof("create new processor with option %v", clbCtrlOption)
		proc, err := processor.NewProcessor(clbCtrlOption)
		if err != nil {
			blog.Errorf("create processor with option %v failed, err %s", clbCtrlOption, err.Error())
			os.Exit(1)
		}

		blog.Infof("init loadbalancer with name %s, nettype %s", clbCtrlOption.ClbName, clbCtrlOption.NetType)
		err = proc.Init()
		if err != nil {
			blog.Errorf("init loadbalancer with name %s, nettype %s, err %s", clbCtrlOption.ClbName, clbCtrlOption.NetType, err.Error())
			os.Exit(1)
		}

		versionMetric := metric.NewVersionMetric()
		prometheus.MustRegister(proc)
		promMetric := metric.NewPromMetric()
		jsonStatus := metric.NewJSONStatus(proc.GetStatusFunction())
		metrics := metric.NewClbMetric(port)
		metrics.RegisterResource(versionMetric)
		metrics.RegisterResource(promMetric)
		metrics.RegisterResource(jsonStatus)

		go metrics.Run()

		go proc.Run()

		interupt := make(chan os.Signal, 10)
		signal.Notify(interupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM,
			syscall.SIGUSR1, syscall.SIGUSR2)
		for {
			select {
			case <-interupt:
				fmt.Printf("Get signal from system. Exit\n")
				proc.Stop()
				metrics.Close()
				return
			}
		}

	},
}
