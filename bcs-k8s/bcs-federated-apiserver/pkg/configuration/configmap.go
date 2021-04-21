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
	"path/filepath"
	"time"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

const configMapNamespace string = "bcs-system"
const configMapName string = "bcs-federated-apiserver"

type AggregationConfigMapInfo struct {
	bcsStorageAddress         string
	bcsStoragePodUri          string
	bcsStorageToken           string
	memberClusterOverride     string
	memberClusterIgnorePrefix string
}

func (acm *AggregationConfigMapInfo) SetAggregationInfo() {
	config, err := rest.InClusterConfig()
	if err != nil {
		// fallback to kubeconfig
		kubeconfig := filepath.Join("~", ".kube", "config")
		if envvar := os.Getenv("KUBECONFIG"); len(envvar) > 0 {
			kubeconfig = envvar
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			klog.Errorf("The kubeconfig cannot be loaded: %v\n", err)
			os.Exit(1)
		}
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorf("Failed to create clientset: %v\n", err)
		os.Exit(1)
	}

	for {
		BcsStorageAddressConfig, err := clientSet.CoreV1().ConfigMaps(configMapNamespace).Get(context.TODO(),
			configMapName,
			metav1.GetOptions{})
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

func (acm *AggregationConfigMapInfo) GetBcsStorageAddress() string {
	return acm.bcsStorageAddress
}

func (acm *AggregationConfigMapInfo) GetBcsStoragePodUri() string {
	return acm.bcsStoragePodUri
}

func (acm *AggregationConfigMapInfo) GetBcsStorageToken() string {
	return acm.bcsStorageToken
}

func (acm *AggregationConfigMapInfo) GetClusterOverride() string {
	return acm.memberClusterOverride
}

func (acm *AggregationConfigMapInfo) GetClusterIgnorePrefix() string {
	return acm.memberClusterIgnorePrefix
}
