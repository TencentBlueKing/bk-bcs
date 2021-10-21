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
)

//ServerOption is option in flags
type ServerOption struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.CertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig
	conf.LocalConfig
	conf.CustomCertConfig

	VerifyClientTLS bool `json:"verify_client_tls" value:"false" usage:"verify client when brings up a tls server" mapstructure:"verify_client_tls"`

	BKIamAuth AuthOption `json:"bkiam_auth"`

	BKE BKEOptions `json:"bke"`

	Edition string `json:"edition" value:"ieod" usage:"api edition"`

	MesosWebconsoleProxyPort uint `json:"mesos_webconsole_proxy_port" value:"8083" usage:"Port to connect to mesos webconsole proxy"`

	TKE TKEOptions `json:"tke"`

	PeerToken string `json:"peer_token" value:"" usage:"peer token to auth with each other, only used to websocket peer"`
}

// BKEOptions bke options
type BKEOptions struct {
	DSN                        string                     `json:"mysql_dsn" value:"" usage:"dsn for connect to mysql"`
	BootStrapUsers             []BootStrapUser            `json:"bootstrap_users"`
	ClusterCredentialsFixtures CredentialsFixturesOptions `json:"cluster_credentials_fixtures"`

	TurnOnRBAC bool `json:"turn_on_rbac" value:"false" usage:"turn on the rbac"`
	TurnOnAuth bool `json:"turn_on_auth" value:"false" usage:"turn on the auth"`
	TurnOnConf bool `json:"turn_on_conf" value:"false" usage:"turn on the conf"`

	RbacDatas []RbacData `json:"rbac_data"`
}

//TKEOptions tke api operation operation
type TKEOptions struct {
	SecretId  string `json:"secret_id" value:"" usage:"tke user account secret id"`
	SecretKey string `json:"secret_key" value:"" usage:"tke user account secret key"`
	CcsHost   string `json:"ccs_host" value:"" usage:"tke ccs host domain"`
	CcsPath   string `json:"ccs_path" value:"" usage:"tke ccs path"`
}

// RbacData rbac data for specified cluster
type RbacData struct {
	Username  string   `json:"user_name"`
	ClusterId string   `json:"cluster_id"`
	Roles     []string `json:"roles"`
}

//CredentialsFixturesOptions option for enable cluster specified token, deprecated
type CredentialsFixturesOptions struct {
	Enabled     bool         `json:"is_enabled_fixtures_credentials"`
	Credentials []Credential `json:"credentials"`
}

// Credential specified token for cluster, deprecated
type Credential struct {
	ClusterID string `json:"cluster_id"`
	Type      string `json:"type"`
	Server    string `json:"server"`
	CaCert    string `json:"ca_cert"`
	Token     string `json:"token"`
}

//BootStrapUser user for system start up
type BootStrapUser struct {
	Name        string   `json:"name"`
	IsSuperUser bool     `json:"is_super_user"`
	Tokens      []string `json:"tokens"`
}

//AuthOption bkiam auth options
type AuthOption struct {
	Auth          bool `json:"auth" value:"false" usage:"use auth mode or not" mapstructure:"auth"`
	RemoteCheck   bool `json:"remote_check" value:"false" usage:"check auth in remote host or not" mapstructure:"remote_check"`
	SkipNoneToken bool `json:"skip_none_token" value:"false" usage:"skip auth check when token no specified" mapstructure:"skip_none_token"`

	Version string `json:"auth_version" value:"3" usage:"bkiam version, 2 or 3." mapstructure:"auth_version"`

	ApiGwRsaFile string `json:"apigw_rsa_file" value:"" usage:"apigw rsa public key file" mapstructure:"apigw_rsa_file"`

	AuthTokenSyncTime int `json:"auth_token_sync_time" value:"10" usage:"time ticker for syncing token in cache, seconds" mapstructure:"auth_token_sync_time"`

	BKIamAuthHost       string          `json:"bkiam_auth_host" value:"" usage:"bkiam auth server host" mapstructure:"bkiam_auth_host"`
	BKIamAuthAppCode    string          `json:"bkiam_auth_app_code" value:"" usage:"app code for communicating with auth" mapstructure:"bkiam_auth_app_code"`
	BKIamAuthAppSecret  string          `json:"bkiam_auth_app_secret" value:"" usage:"app secret for communicating with auth" mapstructure:"bkiam_auth_app_secret"`
	BKIamAuthSystemID   string          `json:"bkiam_auth_system_id" value:"" usage:"system id in auth service" mapstructure:"bkiam_auth_system_id"`
	BKIamAuthScopeID    string          `json:"bkiam_auth_scope_id" value:"" usage:"scope id in auth service" mapstructure:"bkiam_auth_scope_id"`
	BKIamZookeeper      string          `json:"bkiam_auth_zookeeper" value:"" usage:"zookeeper for auth token storage" mapstructure:"bkiam_auth_zookeeper"`
	BKIamTokenWhiteList []AuthWhitelist `json:"bkiam_auth_token_whitelist" value:"" usage:"token whitelist for bkiam"`
	BKIamAuthSubServer  string          `json:"bkiam_auth_sub_server" value:"" usage:"bkiam auth subserver" mapstructure:"bkiam_auth_sub_server"`
}

// AuthWhitelist white list for bkiam
type AuthWhitelist struct {
	Token string          `json:"token"`
	Scope []AuthWLCluster `json:"scope"`
}

// AuthWLCluster cluster id & namespace for whitelist
type AuthWLCluster struct {
	ClusterID string   `json:"cluster_id"`
	Namespace []string `json:"namespace"`
}

//NewServerOption create a ServerOption object
func NewServerOption() *ServerOption {
	s := ServerOption{}
	return &s
}

//Parse configuration item parsed
func Parse(ops *ServerOption) error {
	conf.Parse(ops)
	return nil
}
