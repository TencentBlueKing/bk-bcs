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

package configuration

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/utils"

	"k8s.io/klog"
	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
)

// AggregationClusterInfo is the struct of final member cluster list.
type AggregationClusterInfo struct {
	memberClusterList string
}

// SetClusterInfo is the function for setting the member clusterList info.
// if the configmap has set the "memberClusterOverride", using it;
// otherwise using the clusterList from the KubeFedCluster resource of the namespace kube-federation-system.
func (aci *AggregationClusterInfo) SetClusterInfo(acm *AggregationConfigMapInfo) {
	if acm.GetClusterOverride() != "" {
		klog.Infoln("Get memberClusterList from AggregationConfigMapInfo of ClusterOverride.")
		aci.memberClusterList = strings.ToUpper(acm.GetClusterOverride())
	} else {
		klog.Infoln("Get memberClusterList from kubeFederated member cluster.")

		clientSet, err := utils.NewKubeFedClientSet()
		if err != nil {
			klog.Errorf("Failed to create clientSet: %v", err)
			os.Exit(1)
		}

		clusterList := &fedv1b1.KubeFedClusterList{}

		for {
			err = clientSet.List(context.TODO(), clusterList, "kube-federation-system")
			if err != nil {
				klog.Warningf("Error retrieving list of federated clusters: %v\n", err)
			} else {
				if len(clusterList.Items) == 0 {
					klog.Errorln("No federated clusters found, wait for join KubeFed member cluster")
				} else {
					aci.memberClusterList = ""
					aci.parseClusterListFromKubeFedClusterList(clusterList, acm)
					break
				}
			}

			klog.Errorln("Can not get the member cluster list from kubeFederated member cluster, " +
				"wait for 30 seconds for next loop")
			time.Sleep(30 * time.Second)
		}
	}
	klog.Infoln("Get memberClusterList: " + aci.memberClusterList)
}


// parseClusterListFromKubeFedClusterList function is to setting the aci.memberClusterList
// from kubeFedClusterList, which is dealt with the action of Upper and IgnorePrefix.
func (aci *AggregationClusterInfo)parseClusterListFromKubeFedClusterList(clusterList *fedv1b1.
	KubeFedClusterList, acm *AggregationConfigMapInfo) {
	for _, cluster := range clusterList.Items {
		var clusterTmp string
		if acm.GetClusterIgnorePrefix() != "" {
			clusterTmp = strings.TrimPrefix(cluster.Name,
				acm.GetClusterIgnorePrefix())
		} else {
			clusterTmp = cluster.Name
		}
		aci.memberClusterList += strings.ToUpper(clusterTmp) + ","
	}
	aci.memberClusterList = strings.TrimRight(aci.memberClusterList, ",")
}


// GetClusterList return the memberClusterList to the public Callers
func (aci *AggregationClusterInfo) GetClusterList() string {
	return aci.memberClusterList
}