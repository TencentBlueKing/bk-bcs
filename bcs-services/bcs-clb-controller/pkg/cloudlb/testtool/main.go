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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"io/ioutil"

	networkv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/network/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/cloudlb/qcloud"
)

var action string
var file1 string
var file2 string
var lbFile string

func checkErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}

func init() {
	blog.InitLogs(
		conf.LogConfig{
			LogDir:       "",
			LogMaxSize:   500,
			LogMaxNum:    10,
			ToStdErr:     true,
			AlsoToStdErr: true,
			Verbosity:    5,
		},
	)
}

func main() {
	flag.StringVar(&action, "a", "descLB", "")
	flag.StringVar(&file1, "f1", "", "")
	flag.StringVar(&file2, "f2", "", "")
	flag.StringVar(&lbFile, "lbf", "./lb.json", "")
	flag.Parse()

	data, err := ioutil.ReadFile(lbFile)
	checkErr(err)
	lbInfo := networkv1.CloudLoadBalancer{}
	err = json.Unmarshal(data, &lbInfo)
	checkErr(err)

	clbClient, err := qcloud.NewClient(&lbInfo)
	checkErr(err)
	err = clbClient.LoadConfig()
	checkErr(err)

	switch action {
	case "queryLB":
		lbInfo, err := clbClient.DescribeLoadbalance(lbInfo.Name)
		checkErr(err)
		bytes, _ := json.Marshal(lbInfo)
		checkErr(err)
		fmt.Println(string(bytes))

	case "createLB":
		lb, err := clbClient.CreateLoadbalance()
		if err != nil {
			fmt.Println(err.Error())
		}
		bytes, err := json.Marshal(lb)
		checkErr(err)
		err = ioutil.WriteFile(lbFile, bytes, 0666)
		checkErr(err)

	case "add":
		bytes, err := ioutil.ReadFile(file1)
		checkErr(err)
		listener := networkv1.CloudListener{}
		err = json.Unmarshal(bytes, &listener)
		checkErr(err)
		err = clbClient.Add(&listener)
		checkErr(err)
		bytes, err = json.Marshal(&listener)
		checkErr(err)
		ioutil.WriteFile(file1, bytes, 0666)

	case "update":
		bytes1, err := ioutil.ReadFile(file1)
		checkErr(err)
		bytes2, err := ioutil.ReadFile(file2)
		checkErr(err)
		listener1 := networkv1.CloudListener{}
		err = json.Unmarshal(bytes1, &listener1)
		checkErr(err)
		listener2 := networkv1.CloudListener{}
		err = json.Unmarshal(bytes2, &listener2)
		checkErr(err)
		err = clbClient.Update(&listener1, &listener2)
		checkErr(err)
		bytes, err := json.Marshal(&listener2)
		checkErr(err)
		ioutil.WriteFile(file2, bytes, 0666)

	case "delete":
		bytes, err := ioutil.ReadFile(file1)
		checkErr(err)
		listener := networkv1.CloudListener{}
		err = json.Unmarshal(bytes, &listener)
		checkErr(err)
		err = clbClient.Delete(&listener)
		checkErr(err)
	}
}
