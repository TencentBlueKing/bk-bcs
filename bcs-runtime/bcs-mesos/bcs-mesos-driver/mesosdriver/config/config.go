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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
)

//CertConfig is configuration of Cert
type CertConfig struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
	IsSSL      bool
}

//MesosDriverConfig is a configuration of mesos driver
type MesosDriverConfig struct {
	Address      string
	Port         uint
	ExternalIp   string
	ExternalPort uint

	MetricPort uint

	RegDiscvSvr   string
	SchedDiscvSvr string
	Cluster       string

	ServCert   *CertConfig
	ClientCert *CertConfig

	AdmissionWebhook bool
	//KubeConfig kubeconfig for CustomResource
	KubeConfig string
	//MesosWebconsoleProxyPort
	MesosWebconsoleProxyPort uint

	// websocket register
	RegisterWithWebsocket bool
	RegisterToken         string
	RegisterURL           string
	InsecureSkipVerify    bool

	Etcd *registry.CMDOptions
}
