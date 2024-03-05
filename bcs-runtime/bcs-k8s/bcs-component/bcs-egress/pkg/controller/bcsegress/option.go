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

package bcsegress

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	// init all flags
	viper.SetEnvPrefix("BCS")
	pflag.String("namespace", "bcs-system", "namespace that BCSEgress controller is running in")
	viper.BindEnv("namespace") // nolint
	pflag.String("name", "egress-controller", "name that BCSEgress controller is running in with")
	viper.BindEnv("name") // nolint
	pflag.String("tamplate", "./template/nginx-template.conf", "tamplate use for proxy configuration generation")
	viper.BindEnv("tamplate") // nolint
	pflag.String("generate_dir", "./generate/", "directory for configuration generating")
	viper.BindEnv("generate_dir") // nolint
	viper.BindPFlags(pflag.CommandLine) // nolint
}

// NewOptionFromFlagAndEnv create option from env or command line
func NewOptionFromFlagAndEnv() *EgressOption {
	egress := &EgressOption{
		Namespace:    viper.GetString("namespace"),
		Name:         viper.GetString("name"),
		TemplateFile: viper.GetString("template"),
		GenerateDir:  viper.GetString("generate_dir"),
	}
	return egress
}

// EgressOption all options that required for BCSEgressController
type EgressOption struct {
	Namespace       string
	Name            string
	TemplateFile    string
	GenerateDir     string
	ProxyExecutable string
	ProxyConfig     string
}
