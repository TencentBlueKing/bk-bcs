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
	registry "github.com/Tencent/bk-bcs/bcs-common/pkg/registryv4"
)

// UserManagerOptions cmd option for user-manager
// NOCC:golint/lll(设计如此:)
// nolint
type UserManagerOptions struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.CertConfig
	conf.LogConfig
	conf.LocalConfig
	conf.ProcessConfig
	JWTKeyConfig

	VerifyClientTLS bool            `json:"verify_client_tls" value:"false" usage:"verify client when brings up a tls server" mapstructure:"verify_client_tls"`
	RedisDSN        string          `json:"redis_dsn" value:"" usage:"dsn for connect to redis"`
	DSN             string          `json:"mysql_dsn" value:"" usage:"dsn for connect to mysql"`
	BootStrapUsers  []BootStrapUser `json:"bootstrap_users"`
	TKE             TKEOptions      `json:"tke"`
	PeerToken       string          `json:"peer_token" value:"" usage:"peer token to authorize with each other, only used to websocket peer"`
	// go-micro etcd registry feature support
	Etcd         registry.CMDOptions `json:"etcdRegistry"`
	InsecureEtcd bool                `json:"insecure_etcd" value:"false" usage:"if true, will use insecure etcd registry"`
	// token notify feature
	TokenNotify TokenNotifyOptions `json:"token_notify"`

	IAMConfig        IAMConfig   `json:"iam_config"`
	PermissionSwitch bool        `json:"permission_switch"`
	Cmdb             CmdbConfig  `json:"cmdb"`
	CommunityEdition bool        `json:"community_edition"`
	TracingConf      TracingConf `json:"tracing_conf"`
	BcsAPI           BcsAPI      `json:"bcs_api"`
	Encrypt          Encrypt     `json:"encrypt" yaml:"encrypt"`
	Activity         Activity    `json:"activity" yaml:"activity"`
}

// TracingConf tracing config
type TracingConf struct {
	Enabled       bool              `json:"enabled" usage:"enable trace"`
	Endpoint      string            `json:"endpoint" usage:"Collector service endpoint"`
	Token         string            `json:"token" usage:"token for collector service"`
	ResourceAttrs map[string]string `json:"resource_attrs" usage:"attributes of traced service"`
}

// IAMConfig iam config
type IAMConfig struct {
	SystemID  string `json:"system_id"`
	AppCode   string `json:"app_code"`
	AppSecret string `json:"app_secret"`

	External    bool   `json:"external"`
	GateWayHost string `json:"gateWay_host"`
	IAMHost     string `json:"iam_host"`
	BkiIAMHost  string `json:"bki_iam_host"`

	Metric      bool `json:"metric"`
	ServerDebug bool `json:"server_debug"`
}

// TKEOptions tke api option
type TKEOptions struct {
	SecretID  string `json:"secret_id" value:"" usage:"tke user account secret id"`
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

// TokenNotifyOptions token notify option
type TokenNotifyOptions struct {
	Feature      bool      `json:"feature" value:"false" usage:"if true, will enable token notify feature"`
	DryRun       bool      `json:"dry_run" value:"false" usage:"if true, will not send notification"`
	NotifyCron   string    `json:"notify_cron" value:"0 10 * * *" usage:"cron expression for notify"`
	EmailTitle   string    `json:"email_title" value:"" usage:"email title"`
	EmailContent string    `json:"email_content" value:"" usage:"email content with html format"`
	RtxTitle     string    `json:"rtx_title" value:"" usage:"rtx title"`
	RtxContent   string    `json:"rtx_content" value:"" usage:"rtx content with format"`
	ESBConfig    ESBConfig `json:"esb_config"`
}

// ESBConfig esb config
type ESBConfig struct {
	AppCode       string `json:"app_code" value:"" usage:"app code"`
	AppSecret     string `json:"app_secret" value:"" usage:"app secret"`
	APIHost       string `json:"api_host" value:"" usage:"api host"`
	SendEmailPath string `json:"send_email_path" value:"/api/c/compapi/v2/cmsi/send_mail/" usage:"send email path"`
	SendRtxPath   string `json:"send_rtx_path" value:"/api/c/compapi/v2/cmsi/send_rtx/" usage:"send rtx path"`
}

// JWTKeyConfig config jwt sign key
// NOCC:golint/lll(设计如此:)
// nolint
type JWTKeyConfig struct {
	JWTPublicKeyFile  string `json:"jwt_public_key_file" value:"" usage:"JWT public key file" mapstructure:"jwt_public_key_file"`
	JWTPrivateKeyFile string `json:"jwt_private_key_file" value:"" usage:"JWT private key file" mapstructure:"jwt_private_key_file"`
}

// CmdbConfig for cmdb
type CmdbConfig struct {
	Enable     bool   `json:"enable"`
	AppCode    string `json:"app_code"`
	AppSecret  string `json:"app_secret"`
	BkUserName string `json:"bk_user_name"`
	Host       string `json:"host"`
}

// BcsAPI bcs api config
type BcsAPI struct {
	Host      string `json:"host" usage:"enable http host"`
	InnerHost string `json:"inner_host" usage:"enable http host"`
	Token     string `json:"token" usage:"token for calling service"`
}

// Encrypt define encrypt config
type Encrypt struct {
	Enable    bool          `json:"enable" yaml:"enable"`
	Algorithm string        `json:"algorithm" yaml:"algorithm"`
	Secret    EncryptSecret `json:"secret" yaml:"secret"`
}

// EncryptSecret define encrypt secret
type EncryptSecret struct {
	Key    string `json:"key" yaml:"key"`
	Secret string `json:"secret" yaml:"secret"`
}

// Activity 操作记录清理
type Activity struct {
	Duration     string   `json:"duration" yaml:"duration" usage:"cleaning time"`
	Interval     string   `json:"interval" yaml:"interval" usage:"timed tasks"`
	ResourceType []string `json:"resource_type" yaml:"resource_type"`
}
