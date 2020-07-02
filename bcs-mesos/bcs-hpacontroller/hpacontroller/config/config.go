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

package config

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
)

//CertConfig is configuration of Cert
type CertConfig struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
	IsSSL      bool
}

type Config struct {
	//bcs mesos cluster zk, list/watch application info
	//example: 127.0.0.1:2181,127.0.0.2:2181,127.0.0.3:2181
	ClusterZkAddr string

	//kubeconfig to connect to bcs mesos kube-api, list/watch application info
	KubeConfig string

	//bcs service zk, to discovery bcs-driver and bcs-storage
	//example: 127.0.0.1:2181,127.0.0.2:2181,127.0.0.3:2181
	BcsZkAddr string

	//bcs cluster id
	ClusterID string

	//container resources cadvisor port
	CadvisorPort int

	//client https certs
	ClientCert *CertConfig `json:"-"`

	// The following fields define time interval from which metrics were
	// collected from the interval [interval duration].
	// second, default 30s
	CollectMetricsInterval int

	// the latest value of metrics as an average aggregated across window seconds
	// second, default 30
	CollectMetricsWindow int

	// The period for which autoscaler will look backwards and
	// not scale down below any recommendation it made during that period
	// second, default 300
	DownscaleStabilization int64

	// The period for which autoscaler will look backwards and
	// not scale up below any recommendation it made during that period
	// second, default 180
	UpscaleStabilization int64

	// The period for syncing the autoscaler metrics
	// second, default 10s
	MetricsSyncPeriod int

	// The minimum change (from 1.0) in the desired-to-actual metrics ratio
	// for the autoscaler to consider scaling
	// if metric current/metric target > AutoscalerTolerance or < AutoscalerTolerance, then scale
	// float, default 0.2
	AutoscalerTolerance float32
}

//NewConfig create a config object
func NewConfig() *Config {
	return &Config{
		CollectMetricsInterval: 30,
		CollectMetricsWindow:   30,
		DownscaleStabilization: 300,
		UpscaleStabilization:   180,
		MetricsSyncPeriod:      10,
		AutoscalerTolerance:    0.2,
		ClientCert: &CertConfig{
			CertPasswd: static.ClientCertPwd,
			IsSSL:      false,
		},
	}
}
