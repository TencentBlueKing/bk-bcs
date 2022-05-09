/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package options

// LogConfig option for log
type LogConfig struct {
	LogDir          string `json:"dir"`
	LogMaxSize      uint64 `json:"maxsize"`
	LogMaxNum       int    `json:"maxnum"`
	ToStdErr        bool   `json:"tostderr"`
	AlsoToStdErr    bool   `json:"alsotostderr"`
	Verbosity       int32  `json:"v"`
	StdErrThreshold string `json:"stderrthreshold"`
	VModule         string `json:"vmodule"`
	TraceLocation   string `json:"backtraceat"`
}

// PluginImage option for plugins' image
type PluginImage struct {
	Registry   string `json:"registry"`
	PullPolicy string `json:"pullpolicy"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}

// Plugin option for plugins
type Plugin struct {
	ServerImage PluginImage `json:"serverimage"`
	ClientImage PluginImage `json:"clientimage"`
}

// ArgocdControllerOptions options of bcs argocd server
type ArgocdControllerOptions struct {
	Debug      bool      `json:"debug"`
	KubeConfig string    `json:"kubeconfig"`
	MasterURL  string    `json:"masterurl"`
	Plugin     Plugin    `json:"plugin"`
	BcsLog     LogConfig `json:"bcslog"`
}
