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

// Package systemappcheck xxx
package systemappcheck

import (
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// CheckTKENetwork xxx
func CheckTKENetwork(cluster *plugin_manager.ClusterConfig) []plugin_manager.CheckItem {
	result := make([]plugin_manager.CheckItem, 0, 0)

	cm, err := cluster.ClientSet.CoreV1().ConfigMaps("kube-system").Get(util.GetCtx(10*time.Second), "tke-network-conf", metav1.GetOptions{ResourceVersion: "0"})
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			klog.Errorf("%s %s", cluster.ClusterID, err.Error())
		}
		return result
	}

	tkeNetworkConfig := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(cm.Data["tke-network-conf.yaml"]), &tkeNetworkConfig)
	if err != nil {
		klog.Errorf("%s %s", cluster.ClusterID, err.Error())
		return result
	}

	if cidrList, ok := tkeNetworkConfig["cidr.cluster-cidrs"].([]interface{}); ok {
		for _, cidr := range cidrList {
			cluster.Cidr = append(cluster.Cidr, cidr.(string))
		}

	}

	return result
}
