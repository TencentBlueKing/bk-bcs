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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/conf"
	"bk-bcs/bcs-common/pkg/esb/cmdbv3"
)

func init() {
	pflag.String("host", "", "")
	pflag.String("command", "", "")
	pflag.Int64("biz", 0, "")
	pflag.String("config", "", "")
	pflag.String("cluster", "", "")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	blog.InitLogs(conf.LogConfig{
		Verbosity:       5,
		LogDir:          "./logs",
		LogMaxSize:      500,
		LogMaxNum:       20,
		StdErrThreshold: "2",
		AlsoToStdErr:    true,
	})
}

func loadFile(file string) []byte {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return data
}

func main() {

	client := cmdbv3.NewClientInterface(viper.GetString("host"), nil)
	client.SetDefaultHeader(http.Header{
		"Host":                      []string{"cmdb.test.com"},
		"HTTP_BLUEKING_SUPPLIER_ID": []string{"0"},
		"BK_User":                   []string{"admin"},
	})
	command := viper.GetString("command")
	switch command {
	case "create":
		param := new(cmdbv3.CreatePod)
		data := loadFile(viper.GetString("config"))
		err := json.Unmarshal(data, param)
		if err != nil {
			panic(err)
		}
		re, err := client.CreatePod(viper.GetInt64("biz"), param)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%#v\n", re)
	case "update":
		param := new(cmdbv3.UpdatePod)
		data := loadFile(viper.GetString("config"))
		err := json.Unmarshal(data, param)
		if err != nil {
			panic(err)
		}
		re, err := client.UpdatePod(viper.GetInt64("biz"), param)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%#v\n", re)
	case "delete":
		param := new(cmdbv3.DeletePod)
		data := loadFile(viper.GetString("config"))
		err := json.Unmarshal(data, param)
		if err != nil {
			panic(err)
		}
		re, err := client.DeletePod(viper.GetInt64("biz"), param)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%#v\n", re)
	case "list":
		re, err := client.ListClusterPods(viper.GetInt64("biz"), viper.GetString("cluster"))
		if err != nil {
			panic(err)
		}
		fmt.Printf("%#v\n", re.Data)
	case "listtopo":
		re, err := client.SearchBusinessTopoWithStatistics(viper.GetInt64("biz"))
		if err != nil {
			panic(err)
		}
		fmt.Printf("%#v\n", re)
		str, _ := json.Marshal(re)
		fmt.Printf("%s", str)
	}

}
