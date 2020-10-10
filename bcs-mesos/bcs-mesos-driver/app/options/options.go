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

package options

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-driver/mesosdriver/config"
)

// MesosDriverOptionsOut mesos driver command line options
type MesosDriverOptionsOut struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.CertConfig
	conf.LicenseServerConfig

	conf.LogConfig
	conf.ProcessConfig
	SchedDiscvSvr            string `json:"sched_regdiscv" value:"127.0.0.1:2181" usage:"the address to discove schdulers"`
	Cluster                  string `json:"cluster" value:"" usage:"the cluster ID under bcs"`
	AdmissionWebhook         bool   `json:"admission_webhook" value:"false" usage:"whether admission webhook"`
	KubeConfig               string `json:"kubeconfig" value:"" usage:"kube config for custom resource feature and etcd storage"`
	MesosWebconsoleProxyPort uint   `json:"mesos_webconsole_proxy_port" value:"8083" usage:"Port to connect to mesos webconsole proxy"`
	// websocket register
	RegisterWithWebsocket bool   `json:"register-with-websocket" value:"false" usage:"whether register to bcs-api with websocket"`
	RegisterToken         string `json:"register-token" value:"" usage:"register token to register to bcs-api"`
	RegisterURL           string `json:"register-url" value:"" usage:"bcs-api url to register"`
	InsecureSkipVerify    bool   `json:"insecure-skip-verify" value:"false" usage:"whether insecure skip verify"`

	Etcd registry.CMDOptions `json:"etcdRegistry"`
}

//MesosDriverOption is option in flags
type MesosDriverOption struct {
	DriverConf *config.MesosDriverConfig
}

//NewMesosDriverOption create MesosDriverOption object
func NewMesosDriverOption(opOut *MesosDriverOptionsOut) *MesosDriverOption {

	return &MesosDriverOption{
		DriverConf: &config.MesosDriverConfig{
			MetricPort:               opOut.MetricPort,
			Address:                  opOut.Address,
			Port:                     opOut.Port,
			ExternalPort:             opOut.ExternalPort,
			ExternalIp:               opOut.ExternalIp,
			Cluster:                  opOut.Cluster,
			RegDiscvSvr:              opOut.BCSZk,
			SchedDiscvSvr:            opOut.SchedDiscvSvr,
			AdmissionWebhook:         opOut.AdmissionWebhook,
			KubeConfig:               opOut.KubeConfig,
			MesosWebconsoleProxyPort: opOut.MesosWebconsoleProxyPort,
			ServCert: &config.CertConfig{
				CAFile:     opOut.CAFile,
				CertFile:   opOut.ServerCertFile,
				KeyFile:    opOut.ServerKeyFile,
				CertPasswd: static.ServerCertPwd,
				IsSSL:      false,
			},

			ClientCert: &config.CertConfig{
				CAFile:     opOut.CAFile,
				CertFile:   opOut.ClientCertFile,
				KeyFile:    opOut.ClientKeyFile,
				CertPasswd: static.ClientCertPwd,
				IsSSL:      false,
			},
			RegisterWithWebsocket: opOut.RegisterWithWebsocket,
			RegisterToken:         opOut.RegisterToken,
			RegisterURL:           opOut.RegisterURL,
			InsecureSkipVerify:    opOut.InsecureSkipVerify,
			Etcd:                  &opOut.Etcd,
		},
	}
}
