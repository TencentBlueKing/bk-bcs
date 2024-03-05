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

// Package config xxx
package config

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
)

// CertConfig is configuration of Cert
type CertConfig struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
	IsSSL      bool
}

// CustomSchedulerConfig is a configuration of CustomScheduler
type CustomSchedulerConfig struct {
	Address         string
	Port            uint
	InsecureAddress string
	InsecurePort    uint
	Sock            string
	MetricPort      uint
	ZkHosts         string
	ServCert        *CertConfig
	ClientCert      *CertConfig

	Cluster                  string
	KubeConfig               string
	KubeMaster               string
	UpdatePeriod             uint
	CustomSchedulerType      string
	VerifyClientTLS          bool
	CniAnnotationKey         string
	CniAnnotationValue       string
	FixedIpAnnotationKey     string
	FixedIpAnnotationValue   string
	CloudNetserviceEndpoints []string
	CloudNetserviceCert      *CertConfig
}

// NewCustomSchedulerConfig create a config object
func NewCustomSchedulerConfig() *CustomSchedulerConfig {
	return &CustomSchedulerConfig{
		Address: "127.0.0.1",
		Port:    80,
		ServCert: &CertConfig{
			CertPasswd: static.ServerCertPwd,
			IsSSL:      false,
		},
		ClientCert: &CertConfig{
			CertPasswd: static.ClientCertPwd,
			IsSSL:      false,
		},
	}
}
