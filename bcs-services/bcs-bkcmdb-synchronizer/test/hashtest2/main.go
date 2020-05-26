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

	jump "github.com/lithammer/go-jump-consistent-hash"
)

func main() {
	users := []string{"BCS-K8S-25085",
		"100.125.224.0",
		"BCS-K8S-25088",
		"BCS-K8S-25089",
		"10.117.231.0",
		"BCS-MESOS-30028",
		"BCS-MESOS-30025",
		"10.117.228.0",
		"100.125.192.0",
		"100.125.202.0",
		"BCS-K8S-40106",
		"BCS-MESOS-30021",
		"10.178.44.0",
		"BCS-K8S-40058",
		"BCS-MESOS-30023",
		"10.178.45.0",
		"10.117.227.0",
		"BCS-MESOSLOLCD-30008",
		"10.222.16.0",
		"10.178.49.0",
		"10.178.50.0",
		"100.125.208.0",
		"BCS-K8S-25098",
		"100.125.200.0",
		"BCS-K8S-25099",
		"100.125.164.0",
		"BCS-K8S-40026",
		"BCS-K8S-40063",
		"BCS-K8S-40062",
		"10.117.229.0",
		"BCS-MESOS-30014",
		"BCS-MESOS-30012",
		"100.125.206.0",
		"100.125.194.0",
		"100.125.162.0",
		"100.125.196.0",
		"100.125.212.0",
		"BCS-K8S-40038",
		"10.230.48.0",
		"100.125.214.0",
		"100.125.198.0",
		"100.125.160.0",
		"BCS-MESOSKPGSPRE-30004",
		"10.211.16.0",
		"100.125.216.0",
		"10.178.51.0",
		"100.125.228.0",
		"100.125.218.0",
		"BCS-K8S-40048",
		"10.178.52.0",
		"10.178.46.0",
		"BCS-MESOSLOLDG001-30002",
		"BCS-K8S-40045",
		"10.117.234.0",
		"BCS-MESOSKPGSAPPLE-30003",
		"BCS-MESOSSHLOL001-30006",
		"10.117.232.0",
		"BCS-MESOSLOLTJ-30007",
		"10.117.230.0",
		"10.178.47.0",
		"10.117.233.0",
		"BCS-K8S-40062˜˜",
		"100.125.204.0",
		"100.125.220.0",
		"BCS-K8S-25128",
		"10.117.235.0"}
	tmpMap := make(map[int][]string)
	tmpMap[0] = make([]string, 0)
	tmpMap[1] = make([]string, 0)
	tmpMap[2] = make([]string, 0)
	tmpMap[3] = make([]string, 0)
	tmpMap[4] = make([]string, 0)
	hasher := jump.New(5, jump.NewCRC64())
	for _, user := range users {
		index := hasher.Hash(user)
		tmpMap[index] = append(tmpMap[index], user)
	}
	for key, arr := range tmpMap {
		fmt.Printf("server %d, len %d %#v\n", key, len(arr), arr)
	}
}
