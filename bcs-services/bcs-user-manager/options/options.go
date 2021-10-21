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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
)

// UserManagerOptions cmd option for user-manager
type UserManagerOptions struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.CertConfig
	conf.LogConfig
	conf.LocalConfig
	conf.ProcessConfig

	VerifyClientTLS bool            `json:"verify_client_tls" value:"false" usage:"verify client when brings up a tls server" mapstructure:"verify_client_tls"`
	DSN             string          `json:"mysql_dsn" value:"" usage:"dsn for connect to mysql"`
	BootStrapUsers  []BootStrapUser `json:"bootstrap_users"`
	TKE             TKEOptions      `json:"tke"`
	PeerToken       string          `json:"peer_token" value:"" usage:"peer token to authorize with each other, only used to websocket peer"`
	//go-micro etcd registry feature support
	Etcd registry.CMDOptions `json:"etcdRegistry"`
}

//TKEOptions tke api option
type TKEOptions struct {
	SecretId  string `json:"secret_id" value:"" usage:"tke user account secret id"`
	SecretKey string `json:"secret_key" value:"" usage:"tke user account secret key"`
	CcsHost   string `json:"ccs_host" value:"" usage:"tke ccs host domain"`
	CcsPath   string `json:"ccs_path" value:"" usage:"tke ccs path"`
}

// BootStrapUser system admin user
type BootStrapUser struct {
	Name     string `json:"name"`
	UserType string `json:"user_type" usage:"optional type: admin, saas, plain"`
	Token    string `json:"token"`
}
