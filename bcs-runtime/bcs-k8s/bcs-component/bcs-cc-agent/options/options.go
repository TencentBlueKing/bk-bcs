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
	conf.MetricConfig
	conf.LogConfig
	conf.ProcessConfig

	EngineType string `json:"engine_type" value:"kubernetes" usage:"the platform that bcs-webhook-server runs in, kubernetes or mesos"`
	KubeConfig string `json:"kubeconfig" value:"" usage:"kubeconfig for kube-apiserver, Only required if out-of-cluster."`
	KubeMaster string `json:"kube-master" value:"" usage:"The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster." mapstructure:"kube-master"`
	EsbUrl     string `json:"esb_url" value:"" usage:"esb api url to privilege"`
	AppCode    string `json:"app_code" value:"" usage:"app_code to call esb cc api"`
	AppSecret  string `json:"app_secret" value:"" usage:"app_secret to call esb cc api"`
	BkUsername string `json:"bk_username" value:"" usage:"bk username to call esb cc api"`
}

//NewServerOption create a ServerOption object
func NewServerOption() *ServerOption {
	s := ServerOption{}
	return &s
}

func Parse(ops *ServerOption) error {
	conf.Parse(ops)
	return nil
}
