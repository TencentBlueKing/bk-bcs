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

package template

import (
	"strings"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func getClusterMasterIPs(cluster *proto.Cluster) string {
	masterIPs := make([]string, 0)
	for ip := range cluster.Master {
		masterIPs = append(masterIPs, ip)
	}

	return strings.Join(masterIPs, ",")
}

func getMasterDomain(cls *proto.Cluster) string {
	server, ok := cls.ExtraInfo[apiServer]
	if ok {
		return server
	}

	return ""
}

func getEtcdDomain(cls *proto.Cluster) string {
	etcd, ok := cls.ExtraInfo[etcdServer]
	if ok {
		return etcd
	}

	return ""
}

func getClusterType(cls *proto.Cluster) string {
	if len(cls.GetExtraClusterID()) > 0 {
		return "1"
	}

	return "0"
}
