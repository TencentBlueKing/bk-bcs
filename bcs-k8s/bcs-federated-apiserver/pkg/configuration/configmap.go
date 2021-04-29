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

// Package configuration offers functions and variables of the bcsStorageInfo and memberClusterList from the
// configmap and the kubeFedCluster resource, which provides the data for RESTFUL apiServer.
package configuration

import (
	"context"
	"os"
	"time"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/utils"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const configMapNamespace string = "bcs-system"
const configMapName      string = "bcs-federated-apiserver"

// AggregationConfigMapInfo is the configmap info from the namespace bcs-system,
// name bcs-federated-apiserver.
type AggregationConfigMapInfo struct {
	bcsStorageAddress         string
	bcsStoragePodUri          string
	bcsStorageToken           string
	memberClusterOverride     string
	memberClusterIgnorePrefix string
}

// SetAggregationInfo gets the configmap of namespace bcs-system, name bcs-federated-apiserver,
// to the struct AggregationConfigMapInfo.
func (acm *AggregationConfigMapInfo) SetAggregationInfo() {

	clientSet, err := utils.NewK8sClientSet()
	if err != nil {
		klog.Errorf("Failed to create clientset: %v\n", err)
		os.Exit(1)
	}

	// Loop parsing the configmap info to the struct, until the correct info is put.
	for {
		BcsStorageAddressConfig, err := clientSet.CoreV1().ConfigMaps(configMapNamespace).Get(context.TODO(),
			configMapName, metav1.GetOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				klog.Warningf("failed to query configmap: %v\n", err)
			} else if kerrors.IsUnauthorized(err) {
				klog.Errorf("Unauthorized to query configmap: %v\n", err)
			} else {
				klog.Errorf("failed to query kubeadm's configmap: %v\n", err)
			}

			klog.Errorln("can not get the bcs-strorage configmap enough info, wait for 30 seconds for next loop")
			time.Sleep(30 * time.Second)
		} else {
			for key, value := range BcsStorageAddressConfig.Data {
				switch key {
				case "bcs-storage-address":
					acm.bcsStorageAddress = value
				case "bcs-storage-pod-uri":
					acm.bcsStoragePodUri = value
				case "bcs-storage-token":
					acm.bcsStorageToken = value
				case "member-cluster-override":
					acm.memberClusterOverride = value
				case "member-cluster-ignore-prefix":
					acm.memberClusterIgnorePrefix = value
				default:
					klog.Errorln("no need to parse it: ", key, value)
				}
			}

			if acm.bcsStorageAddress == "" || acm.bcsStoragePodUri == "" {
				klog.Errorln("bcs-storage address or podURI is null, please check your configmap, " +
					"wait for 30 seconds for next loop")
				time.Sleep(30 * time.Second)
				continue
			}

			klog.Infof("AggregationConfigMapInfo: %+v\n", acm)
			break
		}
	}
}

// GetBcsStorageAddress return the bcsStorageAddress info
func (acm *AggregationConfigMapInfo) GetBcsStorageAddress() string {
	return acm.bcsStorageAddress
}

// GetBcsStoragePodUri return the bcsStoragePodUri info
func (acm *AggregationConfigMapInfo) GetBcsStoragePodUri() string {
	return acm.bcsStoragePodUri
}

// GetBcsStorageToken return the bcsStorageToken info
func (acm *AggregationConfigMapInfo) GetBcsStorageToken() string {
	return acm.bcsStorageToken
}

// GetClusterOverride return the memberClusterOverride info
func (acm *AggregationConfigMapInfo) GetClusterOverride() string {
	return acm.memberClusterOverride
}

// GetClusterIgnorePrefix return the memberClusterIgnorePrefix info
func (acm *AggregationConfigMapInfo) GetClusterIgnorePrefix() string {
	return acm.memberClusterIgnorePrefix
}
