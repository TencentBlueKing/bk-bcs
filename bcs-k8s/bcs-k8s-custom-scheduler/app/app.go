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
	"net/http"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/app/custom-scheduler"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/config"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/options"
)

//Run the customScheduler
func Run(conf *config.CustomSchedulerConfig) {
	customSched := custom_scheduler.NewCustomScheduler(conf)
	//start customSched, and http service
	err := customSched.Start()
	if err != nil {
		blog.Errorf("start processor error %s, and exit", err.Error())
		os.Exit(1)
	}

	return
}

//RunPrometheusMetrics starting prometheus metrics handler
func RunPrometheusMetricsServer(conf *config.CustomSchedulerConfig) {
	http.Handle("/metrics", promhttp.Handler())
	addr := conf.Address + ":" + strconv.Itoa(int(conf.MetricPort))
	go http.ListenAndServe(addr, nil)
}

func ParseConfig(op *options.ServerOption) *config.CustomSchedulerConfig {
	customSchedulerConfig := config.NewCustomSchedulerConfig()

	customSchedulerConfig.Address = op.Address
	customSchedulerConfig.Port = op.Port
	customSchedulerConfig.MetricPort = op.MetricPort
	customSchedulerConfig.InsecureAddress = op.InsecureAddress
	customSchedulerConfig.InsecurePort = op.InsecurePort
	customSchedulerConfig.ZkHosts = op.BCSZk
	customSchedulerConfig.VerifyClientTLS = op.VerifyClientTLS
	customSchedulerConfig.CustomSchedulerType = op.CustomSchedulerType
	customSchedulerConfig.KubeConfig = op.Kubeconfig
	customSchedulerConfig.KubeMaster = op.KubeMaster
	customSchedulerConfig.UpdatePeriod = op.UpdatePeriod
	customSchedulerConfig.Cluster = op.Cluster
	customSchedulerConfig.CniAnnotationKey = op.CniAnnotationKey
	customSchedulerConfig.FixedIpAnnotationKey = op.FixedIpAnnotationKey

	//server cert directory
	if op.CertConfig.ServerCertFile != "" && op.CertConfig.ServerKeyFile != "" {
		customSchedulerConfig.ServCert.CertFile = op.CertConfig.ServerCertFile
		customSchedulerConfig.ServCert.KeyFile = op.CertConfig.ServerKeyFile
		customSchedulerConfig.ServCert.CAFile = op.CertConfig.CAFile
		customSchedulerConfig.ServCert.IsSSL = true
	}

	//client cert directory
	if op.CertConfig.ClientCertFile != "" && op.CertConfig.ClientKeyFile != "" {
		customSchedulerConfig.ClientCert.CertFile = op.CertConfig.ClientCertFile
		customSchedulerConfig.ClientCert.KeyFile = op.CertConfig.ClientKeyFile
		customSchedulerConfig.ClientCert.CAFile = op.CertConfig.CAFile
		customSchedulerConfig.ClientCert.IsSSL = true
	}

	return customSchedulerConfig
}
