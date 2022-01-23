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

package types

import "time"

// WebSocketConfig is config
type WebSocketConfig struct {
	Height          int
	Width           int
	Privilege       bool
	Cmd             []string
	Tty             bool
	WebConsoleImage string
	Token           string
	Origin          string
	User            string
	PodName         string
	ProjectsID      string
	ClusterID       string
	SessionID       string
}

// UserPodConfig
type UserPodConfig struct {
	ServiceAccountToken string
	SourceClusterID     string
	HttpsServerAddress  string
	Username            string
	UserToken           string
	PodName             string
	ConfigMapName       string
}

type RespBase struct {
	Code      int         `json:"code"`
	RequestId string      `json:"request_id"`
	Data      interface{} `json:"data,omitempty"`
}

type Permissions struct {
	Test   bool `json:"test"`
	Prod   bool `json:"prod"`
	Create bool `json:"create"`
}

type APIResponse struct {
	Result  bool        `json:"result"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// UserPodData 用户的pod数据
type UserPodData struct {
	UserName   string
	ProjectID  string
	ClustersID string
	PodName    string
	SessionID  string
	CrateTime  time.Time
}

// PodCmData pod的configMapData PodCmData
type PodCmData struct {
	ApiVersion     string          `yaml:"apiVersion,omitempty"`
	CurrentContext string          `yaml:"current-context"`
	Kind           string          `yaml:"kind,omitempty"`
	Clusters       []PodCmClusters `yaml:"clusters,omitempty"`
	Contexts       []PodCmContexts `yaml:"contexts,omitempty"`
	Users          []PodCmUsers    `yaml:"users,omitempty"`
}

type PodCmCluster struct {
	CertificateAuthority  string `yaml:"certificate-authority,omitempty"`
	Server                string `yaml:"server,omitempty"`
	InsecureSkipTlsVerify bool   `yaml:"insecure-skip-tls-verify,omitempty"`
}

type PodCmClusters struct {
	Cluster PodCmCluster `yaml:"cluster,omitempty"`
	Name    string       `yaml:"name,omitempty"`
}

type PodCmContext struct {
	Cluster   string `yaml:"cluster,omitempty"`
	User      string `yaml:"user,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
}

type PodCmContexts struct {
	Name    string       `yaml:"name,omitempty"`
	Context PodCmContext `yaml:"context,omitempty"`
}

type PodCmUser struct {
	Token string `yaml:"token,omitempty"`
}

type PodCmUsers struct {
	Name string    `yaml:"name,omitempty"`
	User PodCmUser `yaml:"user,omitempty"`
}

// XtermMessage web终端发来的包
type XtermMessage struct {
	MsgType string `json:"type"`   // 类型:resize客户端调整终端, input客户端输入
	Input   string `json:"input"`  // msgtype=input情况下使用
	Rows    uint16 `json:"rows"`   // msgtype=resize情况下使用
	Cols    uint16 `json:"cols"`   // msgtype=resize情况下使用
	Output  string `json:"output"` // 输出
}

// AuditRecord 审计记录
type AuditRecord struct {
	InputRecord  string      `json:"input_record"`
	OutputRecord string      `json:"output_record"`
	SessionID    string      `json:"session_id"`
	Context      interface{} `json:"context"` // 这里使用户信息
	ProjectID    string      `json:"project_id"`
	ClusterID    string      `json:"cluster_id"`
	UserPodName  string      `json:"user_pod_name"`
	Username     string      `json:"username"`
}
