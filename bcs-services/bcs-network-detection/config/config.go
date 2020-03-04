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

//CertConfig is configuration of Cert
type CertConfig struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
	IsSSL      bool
}

type Config struct {
	//address, exmaple: 127.0.0.1
	Address string
	//port
	Port uint
	//bcs zookeeper address
	BcsZk string
	//deploy detection node cluster list
	//example: BCS-MESOS-10000,BCS-MESOS-10001,BCS-MESOS-10002...
	Clusters string
	//esb app code
	AppCode string
	//esb app secret
	AppSecret string
	//esb operator
	Operator string
	//esb url
	EsbUrl string
	//cmdb app id
	AppId int
	//http client cert config
	ClientCert *CertConfig
	//http server cert config
	ServerCert *CertConfig
	//deployment template json file path
	Template string
}
