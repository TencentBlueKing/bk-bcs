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

package config

type Config struct {
	DockerSock   string
	LogbeatDir   string
	TemplateFile string
	PrefixFile   string
	//kube-apiserver config file path
	Kubeconfig string
	// whether to enable remove symbol link in the log path
	// this should be false if deployed as in-cluster mode
	EvalSymlink bool
	// logbeat PID file path
	LogbeatPIDFilePath  string
	NeedReload          bool
	FileExtension       string
	LogbeatOutputFormat string
}

//NewConfig create a config object
func NewConfig() *Config {
	return &Config{}
}
