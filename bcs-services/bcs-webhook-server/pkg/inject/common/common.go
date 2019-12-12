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
	NamespaceBcs                  = "bcs-system"
	StandardConfigType            = "standard"
	BcsSystemConfigType           = "bcs-system"
	DataIdEnvKey                  = "io_tencent_bcs_app_dataid"
	AppIdEnvKey                   = "io_tencent_bcs_app_appid"
	StdoutEnvKey                  = "io_tencent_bcs_app_stdout"
	LogPathEnvKey                 = "io_tencent_bcs_app_logpath"
	ClusterIdEnvKey               = "io_tencent_bcs_app_cluster"
	NamespaceEnvKey               = "io_tencent_bcs_app_namespace"
	BcsWebhookAnnotationInjectKey = "webhook.inject.bkbcs.tencent.com"
)

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
