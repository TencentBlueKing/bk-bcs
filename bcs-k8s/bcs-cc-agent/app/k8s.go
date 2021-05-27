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

package app

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-cc-agent/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type NodeInfo struct {
	*Properties
	CvmRegion string
	CvmZone   string
}

var k8sCacheInfo *NodeInfo

// synchronizeK8sNodeInfo sync node info from bk-cmdb periodically
func synchronizeK8sNodeInfo(config *config.BcsCcAgentConfig) error {
	cfg, err := clientcmd.BuildConfigFromFlags(config.KubeMaster, config.Kubeconfig)
	if err != nil {
		return fmt.Errorf("error building kube config: %s", err.Error())
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building kubernetes clientset: %s", err.Error())
	}

	// get pod and namespace from env, so these two envs must be set
	podName := os.Getenv("POD_NAME")
	podNamespace := os.Getenv("POD_NAMESPACE")
	if podName == "" || podNamespace == "" {
		return fmt.Errorf("env [POD_NAME] or [POD_NAMESPACE] is empty, so can't get pod self info")
	}

	pod, err := kubeClient.CoreV1().Pods(podNamespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error get pod object from apiserver: %s", err.Error())
	}
	nodeName := pod.Spec.NodeName
	hostIp := pod.Status.HostIP

	// init info cache
	k8sCacheInfo = &NodeInfo{}

	// sync info from bk-cmdb periodically
	go func() {
		ticker := time.NewTicker(time.Duration(1) * time.Minute)
		defer ticker.Stop()
		for {
			blog.Info("starting to synchronize node info...")

			nodeProperties, err := getInfoFromBkCmdb(config, hostIp)
			if err != nil {
				blog.Errorf("error synchronizing node info: %s", err.Error())
				continue
			}

			node, err := kubeClient.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
			if err != nil {
				blog.Errorf("error get node from k8s: %s", err.Error())
				continue
			}

			// currentNodeInfo represents the current node info
			currentNodeInfo := &NodeInfo{
				Properties: nodeProperties,
				CvmRegion:  node.Labels["failure-domain.beta.kubernetes.io/region"],
				CvmZone:    node.Labels["failure-domain.beta.kubernetes.io/zone"],
			}

			// if nodeInfo updated, then update to file and node label
			if !reflect.DeepEqual(*k8sCacheInfo, *currentNodeInfo) {
				k8sCacheInfo = currentNodeInfo
				err := updateK8sNodeInfo(kubeClient, nodeName, k8sCacheInfo)
				if err != nil {
					blog.Errorf("error updating node info to file and node label: %s", err.Error())
					continue
				}
			}

			select {
			case <-ticker.C:
			}
		}
	}()

	return nil
}
