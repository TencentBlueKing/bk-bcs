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

package option

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// NewOption create new webhook option
func NewOption() *IPLimitOption {
	opt := new(IPLimitOption)
	return opt
}

// IPLimitOption option for ip limit webhook
type IPLimitOption struct {
	PodAnnotationKey    string `json:"pod_annotation_key" value:"tke.cloud.tencent.com/networks" usage:"annotation key in pod which should be hook"`
	PodAnnotationValue  string `json:"pod_annotation_value" value:"bcs-qcloud" usage:"annotation value in pod which should be hook"`
	InjectResourceName  string `json:"inject_resource_name" value:"cloud.bkbcs.tencent.com/eip" usage:"extended resource name in pod which should be injected"`
	ValidateWebhookPath string `json:"validating_webhook_path" value:"/v1/validate" usage:"validate url for webhook"`
	MutatingWebhookPath string `json:"mutating_webhook_path" value:"/v1/mutate" usage:"mutate url for webhook"`
	ServerCertFile      string `json:"server_cert_file" value:"/data/bcs/cert/server-cert.pem" usage:"cert file path for webhook"`
	ServerKeyFile       string `json:"server_key_file" value:"/data/bcs/cert/server-key.pem" usage:"key file path for webhook"`

	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.LogConfig
}
