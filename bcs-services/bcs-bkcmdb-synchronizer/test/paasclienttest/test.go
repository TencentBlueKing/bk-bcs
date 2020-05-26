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

package main

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"bk-bcs/bcs-common/pkg/esb/apigateway/paascc"
)

func init() {
	pflag.String("host", "", "")
	pflag.String("env", "", "")
	pflag.String("command", "", "")
	pflag.String("project", "", "")
	pflag.String("appcode", "", "")
	pflag.String("appsecret", "", "")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}

func main() {
	client := paascc.NewClientInterface(viper.GetString("host"), viper.GetString("appcode"), viper.GetString("appsecret"), nil)

	command := viper.GetString("command")
	switch command {
	case "p":
		re, err := client.ListProjects(viper.GetString("env"))
		if err != nil {
			panic(err)
		}
		fmt.Printf("%#v\n", re)
	case "c":
		re, err := client.ListProjectClusters(viper.GetString("env"), viper.GetString("project"))
		if err != nil {
			panic(err)
		}
		fmt.Printf("%#v\n", re)
	}
}
