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

// Package options xx
package options

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
)

var (
	op = new(AnalysisOptions)
)

// AnalysisOptions defines the options of analysis
type AnalysisOptions struct {
	conf.FileConfig
	conf.LogConfig

	Address string `json:"address"`
	Port    int    `json:"port"`

	AppMetricNamespace string `json:"appMetricNamespace"`
	AppMetricName      string `json:"appMetricName"`

	BKMonitorPushUrl    string `json:"bkMonitorPushUrl"`
	BKMonitorPushDataID int64  `json:"bkMonitorPushDataID"`
	BKMonitorPushToken  string `json:"bkMonitorPushToken"`
	BKMonitorGetUrl     string `json:"bkMonitorGetUrl"`
	BKMonitorGetUser    string `json:"bkMonitorGetUser"`
	BKMonitorGetBizID   int64  `json:"bkMonitorGetBizID"`

	BKCCUrl string `json:"bkccUrl"`

	ExternalAnalysisUrl   string `json:"externalAnalysisUrl"`
	ExternalAnalysisToken string `json:"externalAnalysisToken"`
	IsExternal            bool   `json:"-"`

	Auth         AuthConfig                `json:"auth,omitempty"`
	DBConfig     DBConfig                  `json:"dbConfig,omitempty"`
	ArgoConfig   ArgoConfig                `json:"argoConfig,omitempty"`
	SecretConfig common.SecretStoreOptions `json:"secretConfig,omitempty"`
}

// AuthConfig defines the config of auth
type AuthConfig struct {
	AppCode   string `json:"appCode"`
	AppSecret string `json:"appSecret"`
}

// DBConfig defines the config of db
type DBConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Addr     string `json:"addr"`
	Database string `json:"database"`
	LimitQPS int64  `json:"limitQps"`
}

// ArgoConfig defines the configuration of argo
type ArgoConfig struct {
	ArgoService    string `json:"argoService" value:"" usage:"the service address of argo"`
	ArgoUser       string `json:"argoUser" value:"" usage:"the user of argo"`
	ArgoPass       string `json:"argoPass" value:"" usage:"the password of argo"`
	AdminNamespace string `json:"adminNamespace" value:"" usage:"the password of argo"`
}

// GlobalOptions returns the global option object
func GlobalOptions() *AnalysisOptions {
	return op
}

// Parse the config options
func Parse() *AnalysisOptions {
	conf.Parse(op)
	if op.ExternalAnalysisUrl == "" {
		op.IsExternal = true
	}
	return op
}
