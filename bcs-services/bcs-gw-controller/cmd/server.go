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
	"bk-bcs/bcs-services/bcs-gw-controller/pkg/processor"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var port int
var serviceRegistry string
var cluster string
var backendIPType string
var kubeconfig string

// gwZkHosts, gwZkPath, gwBizID is needed for accessing gw concentrator
var gwZkHosts string
var gwZkPath string
var gwBizID string

// updateInterval update interval for processor
var updateInterval int

// syncPeriod sync period for service discovery
var syncPeriod int

// serviceLabelKey, serviceLabelValue are needed for service selection
var serviceLabelKey string
var serviceLabelValue string

// domainLabelKey, proxyPortLabelKey, portLabelkey, pathLabelKey are needed for gw rule
var domainLabelKey string
var proxyPortLabelKey string
var portLabelkey string
var pathLabelKey string

// log configs
var logDir string
var logMaxSize uint64
var logMaxNum int
var toStdErr bool
var alsoToStdErr bool
var verbosity int32
var stdErrThreshold string
var vModule string
var traceLocation string

// tls configs
var caFile string
var serverCertFile string
var serverKeyFile string
var clientCertFile string
var clientKeyFile string

func init() {
	rootCmd.AddCommand(serverCmd)
	// option for clb controller server process
	serverCmd.Flags().IntVar(&port, "port", 18080, "port for clb controller server")
	serverCmd.Flags().StringVar(&serviceRegistry, "serviceRegistry", "kubernetes", "service registry for clb controller, available: [kubernetes, custom, mesos]")
	serverCmd.Flags().StringVar(&cluster, "cluster", "", "lb name for controller")
	serverCmd.Flags().StringVar(&backendIPType, "backendIPType", "", "backend pod ip network type, available: [overlay, underlay]")
	serverCmd.Flags().IntVar(&updateInterval, "updateInterval", 10, "interval for update operations")
	serverCmd.Flags().StringVar(&gwZkHosts, "gwZkHosts", "127.0.0.1:2181", "zk address for gw service")
	serverCmd.Flags().StringVar(&gwZkPath, "gwZkPath", "", "zk path for discovery gw service")
	serverCmd.Flags().StringVar(&gwBizID, "gwBizID", "", "biz id for gw service")
	serverCmd.Flags().StringVar(&serviceLabelKey, "serviceLabelkey", "gw.bkbcs.tencent.com", "label key for select service")
	serverCmd.Flags().StringVar(&serviceLabelValue, "servicelabelValue", "", "label value select service")
	serverCmd.Flags().StringVar(&domainLabelKey, "domainLabelKey", "domain.gw.bkbcs.tencent.com", "label key for gw domain")
	serverCmd.Flags().StringVar(&proxyPortLabelKey, "proxyPortLabelKey", "proxyport.gw.bkbcs.tencent.com", "label key for gw proxy port")
	serverCmd.Flags().StringVar(&portLabelkey, "portLabelKey", "port.gw.bkbcs.tencent.com", "label key for service port")
	serverCmd.Flags().StringVar(&pathLabelKey, "pathLabelKey", "path.gw.bkbcs.tencent.com", "label key for gw path")
	// for kube registry
	serverCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig for access kube-apiserver, if empty, use in-cluster config, default: empty")
	serverCmd.Flags().IntVar(&syncPeriod, "syncPeriod", 30, "period for synchronize k8s services")
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
	// tls config
	serverCmd.Flags().StringVar(&caFile, "ca_file", "", "CA file. If server_cert_file/server_key_file/ca_file are all set, it will set up an HTTPS server required and verified client cert")
	serverCmd.Flags().StringVar(&serverCertFile, "server_cert_file", "", "Server public key file(*.crt). If both server_cert_file and server_key_file are set, it will set up an HTTPS server")
	serverCmd.Flags().StringVar(&serverKeyFile, "server_key_file", "", "Server private key file(*.key). If both server_cert_file and server_key_file are set, it will set up an HTTPS server")
	serverCmd.Flags().StringVar(&clientCertFile, "client_cert_file", "", "Client public key file(*.crt)")
	serverCmd.Flags().StringVar(&clientKeyFile, "client_key_file", "", "Client private key file(*.key)")
}

func validateArgs() bool {
	if port < 0 || port > 65535 {
		blog.Infof("port %d invalid, must be in range [0, 65535]", port)
		return false
	}
	// currently only support [kubernetes, custom, mesos]
	if serviceRegistry != "kubernetes" && serviceRegistry != "custom" && serviceRegistry != "mesos" {
		blog.Errorf("serviceRegistry %s invalid, must be in (kubernetes, custom, mesos)", serviceRegistry)
		return false
	}
	reg, _ := regexp.Compile("[a-zA-Z0-9-\\.]+")
	if !reg.MatchString(cluster) {
		blog.Errorf("cluster %s invalid, must be [a-zA-Z0-9-\\.]+", cluster)
		return false
	}
	if backendIPType != common.BackendIPTypeOverlay && backendIPType != common.BackendIPTypeUnderlay {
		blog.Errorf("backendIPType %s invalid, backendIPType must be in (overlay, underlay)", backendIPType)
		return false
	}
	if len(gwZkHosts) == 0 {
		blog.Errorf("gwZkHosts cannot be empty")
		return false
	}
	if len(gwZkPath) == 0 {
		blog.Errorf("gwZkPath cannot be empty")
		return false
	}
	if len(gwBizID) == 0 {
		blog.Errorf("gwBizID cannot be empty")
		return false
	}
	if len(serviceLabelKey) == 0 || len(serviceLabelValue) == 0 {
		blog.Errorf("serviceLabelKey or serviceLabelValue cannot be empty")
		return false
	}

	reg, _ = regexp.Compile("[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*(\\/([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?")
	if !reg.MatchString(serviceLabelKey) || !reg.MatchString(domainLabelKey) ||
		!reg.MatchString(portLabelkey) || !reg.MatchString(proxyPortLabelKey) {
		blog.Errorf("one of serviceLabelkey %s, domainLabelKey %s, portLabelkey %s, proxyPortLabelKey %s, invalid, must be form of "+
			"[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*(\\/([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?",
			serviceLabelKey, domainLabelKey, portLabelkey, proxyPortLabelKey)
	}
	reg, _ = regexp.Compile("(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?")
	if !reg.MatchString(serviceLabelValue) {
		blog.Errorf("serviceLabelValue %s invalid, must be form of "+
			"(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?", serviceLabelValue)
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

		gwZkHosts = strings.Replace(gwZkHosts, ";", ",", -1)

		opt := &processor.Option{
			Port:            port,
			ServiceRegistry: serviceRegistry,
			Cluster:         cluster,
			BackendIPType:   backendIPType,
			Kubeconfig:      kubeconfig,
			GwZkHosts:       gwZkHosts,
			GwZkPath:        gwZkPath,
			ServiceLabel: map[string]string{
				serviceLabelKey: serviceLabelValue,
			},
			DomainLabelKey:    domainLabelKey,
			ProxyPortLabelKey: proxyPortLabelKey,
			PortLabelKey:      portLabelkey,
			PathLabelKey:      pathLabelKey,
			TLSOption: processor.TLSOption{
				CaFile:         caFile,
				ServerCertFile: serverCertFile,
				ServerKeyFile:  serverKeyFile,
				ClientCertFile: clientCertFile,
				ClientKeyFile:  clientKeyFile,
			},
			GwBizID:      gwBizID,
			UpdatePeriod: updateInterval,
			SyncPeriod:   syncPeriod,
		}

		proc, err := processor.NewProcessor(opt)
		if err != nil {
			blog.Errorf("create processor failed, err %s", err.Error())
			os.Exit(1)
		}

		go proc.Run()

		interupt := make(chan os.Signal, 10)
		signal.Notify(interupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM,
			syscall.SIGUSR1, syscall.SIGUSR2)
		for {
			select {
			case <-interupt:
				fmt.Printf("Get signal from system. Exit\n")
				proc.Stop()
				return
			}
		}
	},
}
