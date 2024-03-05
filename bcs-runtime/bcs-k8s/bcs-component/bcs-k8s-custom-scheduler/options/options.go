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

// Package options xxx
package options

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// ServerOption is option in flags
// nolint
type ServerOption struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.CertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	CustomSchedulerType           string `json:"custom_scheduler_type" value:"" usage:"scheduler type: IpSchedulerV1, IpSchedulerV2, etc." mapstructure:"custom_scheduler_type"`
	VerifyClientTLS               bool   `json:"verify_client_tls" value:"false" usage:"verify client when brings up a tls server" mapstructure:"verify_client_tls"`
	Cluster                       string `json:"cluster" value:"" usage:"k8s cluster name" mapstructure:"cluster"`
	Kubeconfig                    string `json:"kubeconfig" value:"" usage:"Path to a kubeconfig. Only required if out-of-cluster."  mapstructure:"kubeconfig"`
	KubeMaster                    string `json:"kube-master" value:"" usage:"The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster." mapstructure:"kube-master"`
	UpdatePeriod                  uint   `json:"update-period" value:"120" usage:"The period by seconds to update netPool from netService" mapstructure:"update-period"`
	CniAnnotationKey              string `json:"cni_annotation_key" value:"" usage:"cni annotation key for IpSchedulerV2" mapstructure:"cni_annotation_key"`
	CniAnnotationValue            string `json:"cni_annotation_value" value:"" usage:"cni annotation value for IpSchedulerV2" mapstructure:"cni_annotation_value"`
	FixedIpAnnotationKey          string `json:"fixed_ip_annotation_key" value:"" usage:"fixed ip annotation key for IpSchedulerV2" mapstructure:"fixed_ip_annotation_key"`
	FixedIpAnnotationValue        string `json:"fixed_ip_annotation_value" value:"" usage:"fixed ip annotation value for IpSchedulerV2" mapstructure:"fixed_ip_annotation_value"`
	CloudNetserviceEndpoints      string `json:"cloud_netservice_endpoints" value:"" usage:"endpoints of cloud netservice"`
	CloudNetserviceClientCaFile   string `json:"cloud_netservice_client_ca_file" value:"" usage:"client ca file for accessing cloud netservice"`
	CloudNetserviceClientCertFile string `json:"cloud_netservice_client_cert_file" value:"" usage:"client cert file for accessing cloud netservice"`
	CloudNetserviceClientKeyFile  string `json:"cloud_netservice_client_key_file" value:"" usage:"client key file for accessing cloud netservice"`
}

// NewServerOption create a ServerOption object
func NewServerOption() *ServerOption {
	s := ServerOption{}
	return &s
}

// Parse xxx
func Parse(ops *ServerOption) error {
	conf.Parse(ops)
	return nil
}
