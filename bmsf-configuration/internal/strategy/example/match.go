/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"

	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/internal/strategy"
)

func main() {

	handler := strategy.NewHandler(nil)
	match := handler.Matcher()

	labels := make(map[string]string)
	labels["k1"] = "v1"
	labels["k2"] = "v2"

	newStrategy := &strategy.Strategy{
		Appid:      "appid01",
		Clusterids: []string{"clusterid01"},
		Zoneids:    []string{"zoneid01"},
		Dcs:        []string{"dc01"},
		IPs:        []string{"127.0.0.1"},
		Labels:     labels,
	}

	ins := &pbcommon.AppInstance{
		Appid:     "appid01",
		Clusterid: "clusterid01",
		Zoneid:    "zoneid01",
		Dc:        "dc01",
		IP:        "127.0.0.1",
		Labels:    "{\"Labels\":{\"k1\":\"v1\", \"k2\":\"v2\"}}",
	}

	if match(newStrategy, ins) {
		fmt.Println("strategy match!")
	} else {
		fmt.Println("strategy unmatch!")
	}
}
