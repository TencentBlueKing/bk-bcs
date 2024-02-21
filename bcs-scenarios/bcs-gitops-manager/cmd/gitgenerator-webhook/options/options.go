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

// Package options xxx
package options

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// Config defines the config
type Config struct {
	conf.FileConfig
	conf.LogConfig

	ListenAddr string `json:"listen_addr" value:"0.0.0.0" usage:"proxy server listen addr"`
	ListenPort int    `json:"listen_port" value:"8080" usage:"proxy server listen port"`

	TlsCert string `json:"tls_cert" value:"/etc/webhook/certs/cert.pem" usage:"webhook server cert"`
	TlsKey  string `json:"tls_key" value:"/etc/webhook/certs/key.pem" usage:"webhook server key"`

	ArgoService string `json:"argo_service" value:"" usage:"the service address of argo"`
	ArgoUser    string `json:"argo_user" value:"" usage:"the user of argo"`
	ArgoPass    string `json:"argo_pass" value:"" usage:"the password of argo"`
}
