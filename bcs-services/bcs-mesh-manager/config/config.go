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

import "crypto/tls"

// Config all config item for bcs-mesh-manager
type Config struct {
	//IstioOperator Docker Hub
	DockerHub string
	//Istio Operator Charts
	IstioOperatorCharts string
	//IstioOperator cr
	IstioConfiguration string
	//bcs api-gateway address
	ServerAddress string
	//api-gateway usertoken
	UserToken string
	//address
	Address string
	//port, grpc port, http port +1
	Port uint
	//metrics port
	MetricsPort string
	//etcd servers
	EtcdServers string
	//etcd cert file
	EtcdCertFile string
	//etcd key file
	EtcdKeyFile string
	//etcd ca file
	EtcdCaFile string
	//server ca file
	ServerCaFile string
	//server key file
	ServerKeyFile string
	//server cert file
	ServerCertFile string
	//is ssl
	IsSsl bool
	//tls config
	TLSConf *tls.Config
	//kubeconfig
	Kubeconfig string
}
