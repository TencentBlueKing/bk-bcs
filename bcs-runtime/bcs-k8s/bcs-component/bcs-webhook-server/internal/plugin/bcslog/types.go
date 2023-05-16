/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bcslog

import (
	"strings"

	bcsv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/apis/bkbcs/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// BcsLogPluginName plugin name for bcs log
	BcsLogPluginName = "bcslog"

	// NamespaceBcs bcs system namespace
	NamespaceBcs = "bcs-system"
	// DefaultConfigType default log config type
	DefaultConfigType = "default"
	// BcsSystemConfigType bcs system log config type
	BcsSystemConfigType = "bcs-system"
	// CustomConfigType custom log config type
	CustomConfigType = "custom"

	// StdDataIDEnvKey std_dataid of bcs app
	StdDataIDEnvKey = "io_tencent_bcs_app_std_dataid_v2"
	// NonStdDataIDEnvKey non_std_dataid of bcs app
	NonStdDataIDEnvKey = "io_tencent_bcs_app_non_std_dataid_v2"

	// AppIDEnvKey appid of bcs app
	AppIDEnvKey = "io_tencent_bcs_app_appid_v2"
	// StdoutEnvKey log to stdout, true or false
	StdoutEnvKey = "io_tencent_bcs_app_stdout_v2"
	// LogPathEnvKey output path of the log
	LogPathEnvKey = "io_tencent_bcs_app_logpath_v2"
	// ClusterIDEnvKey bcs cluster id
	ClusterIDEnvKey = "io_tencent_bcs_app_cluster_v2"
	// LogTagEnvKey 日志标签
	LogTagEnvKey = "io_tencent_bcs_app_label_v2"
	// NamespaceEnvKey the namespace of app
	NamespaceEnvKey = "io_tencent_bcs_app_namespace_v2"
	// BcsWebhookAnnotationInjectKey bcs webhook annotation inject key
	BcsWebhookAnnotationInjectKey = "webhook.inject.bkbcs.tencent.com"
)

// IgnoredNamespaces namespaces to ignore inject
var IgnoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
	NamespaceBcs,
}

// FindBcsSystemConfigType get the matced bcs-system BcsLogConfig
func FindBcsSystemConfigType(bcsLogConfs []*bcsv1.BcsLogConfig) *bcsv1.BcsLogConfig {
	var matchedLogConf *bcsv1.BcsLogConfig
	for _, logConf := range bcsLogConfs {
		if logConf.Spec.ConfigType == BcsSystemConfigType {
			matchedLogConf = logConf
			break
		}
	}
	return matchedLogConf
}

// FindDefaultConfigType find default config type
func FindDefaultConfigType(bcsLogConfs []*bcsv1.BcsLogConfig) *bcsv1.BcsLogConfig {
	var defaultLogConf *bcsv1.BcsLogConfig
	for _, logConf := range bcsLogConfs {
		if logConf.Spec.ConfigType == DefaultConfigType {
			defaultLogConf = logConf
			break
		}
	}
	return defaultLogConf
}

// FindK8sMatchedConfigType get the matched BcsLogConfig
func FindK8sMatchedConfigType(pod *corev1.Pod, bcsLogConfs []*bcsv1.BcsLogConfig) *bcsv1.BcsLogConfig { // nolint
	if len(pod.OwnerReferences) == 0 {
		return nil
	}

	var matchedLogConf *bcsv1.BcsLogConfig
	for _, logConf := range bcsLogConfs {
		if logConf.Spec.ConfigType == CustomConfigType {
			if pod.OwnerReferences[0].Kind == "ReplicaSet" {
				if strings.ToLower(logConf.Spec.WorkloadType) == strings.ToLower("Deployment") &&
					strings.HasPrefix(pod.OwnerReferences[0].Name, logConf.Spec.WorkloadName) { // nolint
					matchedLogConf = logConf
					break
				}
				continue
			}
			if strings.ToLower(pod.OwnerReferences[0].Kind) == strings.ToLower(logConf.Spec.WorkloadType) &&
				pod.OwnerReferences[0].Name == logConf.Spec.WorkloadName { // nolint
				matchedLogConf = logConf
				break
			}
		}
	}
	return matchedLogConf
}

// FindMesosMatchedConfigType get the matched BcsLogConfig
func FindMesosMatchedConfigType(workloadType, workloadName string,
	bcsLogConfs []*bcsv1.BcsLogConfig) *bcsv1.BcsLogConfig { // nolint
	var matchedLogConf *bcsv1.BcsLogConfig
	for _, logConf := range bcsLogConfs {
		if logConf.Spec.ConfigType == CustomConfigType {
			if strings.ToLower(logConf.Spec.WorkloadType) == workloadType && logConf.Spec.WorkloadName == workloadName {
				// nolint
				matchedLogConf = logConf
				break
			}
		}
	}
	return matchedLogConf
}
