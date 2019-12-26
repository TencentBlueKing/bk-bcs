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

package common

import (
	bcsv2 "bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v2"
	mapset "github.com/deckarep/golang-set"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// bcs system namespace
	NamespaceBcs                  = "bcs-system"
	// default log config type
	StandardConfigType            = "standard"
	// bcs system log config type
	BcsSystemConfigType           = "bcs-system"
	// dataid of bcs app
	DataIdEnvKey                  = "io_tencent_bcs_app_dataid"
	// appid of bcs app
	AppIdEnvKey                   = "io_tencent_bcs_app_appid"
	// log to stdout, true or false
	StdoutEnvKey                  = "io_tencent_bcs_app_stdout"
	// output path of the log
	LogPathEnvKey                 = "io_tencent_bcs_app_logpath"
	// bcs cluster id
	ClusterIdEnvKey               = "io_tencent_bcs_app_cluster"
	// the namespace of app
	NamespaceEnvKey               = "io_tencent_bcs_app_namespace"
	BcsWebhookAnnotationInjectKey = "webhook.inject.bkbcs.tencent.com"
)

// namespaces to ignore inject
var IgnoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
	NamespaceBcs,
}

// FindBcsSystemConfigType get the matced bcs-system BcsLogConfig
func FindBcsSystemConfigType(bcsLogConfs []*bcsv2.BcsLogConfig) *bcsv2.BcsLogConfig {
	var matchedLogConf *bcsv2.BcsLogConfig
	for _, logConf := range bcsLogConfs {
		if logConf.Spec.ConfigType == BcsSystemConfigType {
			matchedLogConf = logConf
			break
		}
	}
	return matchedLogConf
}

// FindMatchedConfigType get the matched BcsLogConfig
func FindMatchedConfigType(name string, bcsLogConfs []*bcsv2.BcsLogConfig) *bcsv2.BcsLogConfig {
	var matchedLogConf *bcsv2.BcsLogConfig
	for _, logConf := range bcsLogConfs {
		if logConf.Spec.ConfigType == StandardConfigType {
			matchedLogConf = logConf
			continue
		}
		containerSet := mapset.NewSet()
		for _, containerName := range logConf.Spec.Containers {
			containerSet.Add(containerName)
		}
		if containerSet.Contains(name) {
			matchedLogConf = logConf
			break
		}
	}
	return matchedLogConf
}
