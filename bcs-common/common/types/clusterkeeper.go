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

package types

type ClusterInfoItem struct {
	ClusterID   string           `json:"clusterid,omitempty"`
	Type        string           `json:"type"`
	ModuleInfos []ModuleInfoItem `json:"modules"`
}

type ModuleInfoItem struct {
	MasterModule []string `json:"mastermoudle"`
	IPS          []string `json:"ips"`
}

func NewKeeperClusterInfoOutput() *KeeperClusterInfoOutput {
	return &KeeperClusterInfoOutput{
		ClusterInfos: make(map[string]ClusterInfoItem),
	}
}

type KeeperClusterInfoOutput struct {
	ClusterInfos map[string]ClusterInfoItem
}

type KeeperDBData struct {
	DBData []*DBDataItem `json:"clusterinfo"`
}

func NewDBDataItem() *DBDataItem {
	return &DBDataItem{
		Detail: make(map[string][]string),
	}
}

type DBDataItem struct {
	Type      string              `json:"type"`
	ClusterID string              `json:"clusterid"`
	Detail    map[string][]string `json:"detail"`
}

func ReqTran2DBFmt(clusterInfo *ClusterInfoItem) (dbData *DBDataItem) {
	dbDataItem := NewDBDataItem()
	dbDataItem.Type = clusterInfo.Type
	dbDataItem.ClusterID = clusterInfo.ClusterID
	for _, ModuleInfo := range clusterInfo.ModuleInfos {
		for _, module := range ModuleInfo.MasterModule {
			ips, ok := dbDataItem.Detail[module]
			if ok {
				for _, newIp := range ModuleInfo.IPS {
					found := false
					for _, ip := range ips {
						if newIp == ip {
							found = true
							break
						}
					}
					if !found {
						ips = append(ips, newIp)
					}
				}
				dbDataItem.Detail[module] = ips
			} else {
				dbDataItem.Detail[module] = ModuleInfo.IPS
			}
		}
	}

	return dbDataItem
}

func AppendItem2DbData(newData, srcData *DBDataItem) (dbData *DBDataItem) {
	for newModule, newIps := range newData.Detail {
		srcIps, ok := srcData.Detail[newModule]
		if ok {
			for _, newIp := range newIps {
				found := false
				for _, srcIp := range srcIps {
					if srcIp == newIp {
						found = true
						break
					}
				}

				if !found {
					srcIps = append(srcIps, newIp)
				}
			}
			srcData.Detail[newModule] = srcIps
		} else {
			srcData.Detail[newModule] = newIps
		}
	}

	return srcData
}

/*
{
	"common_master_module":[
		"bcs-rescheduler",
		"bcs-health",
		"bcs-data-watch",
		"bcs-clusterkeeper",
		"zookeeper"
	],
	"type":"mesos",   #用于查询时候的过滤条件
	"clusterinfo":[
		{
			"clusterid":"BCS-MESOSSELFTEST-10000",
			"type":"kebenutes",   #用于创建时表明集群的类型
			"modules":[
				{
					"ips":[
						"127.0.0.12",
						"127.0.0.13",
						"127.0.0.17"
					]
				}
			]
		},
		{
			"clusterid":"BCS-MESOSSELFTEST-10001",
			"type":"mesos",
			"modules":[
				{
					"mastermoudle":[
						"bcs-rescheduler",
						"bcs-health",
						"bcs-data-watch",
						"bcs-clusterkeeper",
						"zookeeper"
					],
					"ips":[
						"127.0.0.12",
						"127.0.0.13",
						"127.0.0.17"
					]
				},
				{
					"mastermoudle":[
						"bcs-rescheduler",
						"bcs-health",
						"bcs-data-watch",
						"bcs-clusterkeeper"
					],
					"ips":[
						"127.0.0.15",
						"127.0.0.14"
					]
				}
			]
		}
	]
}
*/
