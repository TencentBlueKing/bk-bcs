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

package utils

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"sigs.k8s.io/kubefed/pkg/client/generic"
)

// NewKubeFedClientSet to generate a new KubeFedClientSet
func NewKubeFedClientSet() (generic.Client, error) {
	config, err := newConfig()
	if err != nil {
		klog.Errorf("The kubeConfig cannot be loaded: %v\n", err)
		return nil, err
	}

	clientSet, err := generic.New(config)
	if err != nil {
		klog.Errorf("Failed to create clientSet: %v", err)
		return nil, err
	}
	return clientSet, err
}

// NewK8sClientSet to generate a new K8sClientSet
func NewK8sClientSet() (*kubernetes.Clientset, error) {
	config, err := newConfig()
	if err != nil {
		klog.Errorf("The kubeConfig cannot be loaded: %v\n", err)
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorf("Failed to create clientSet: %v\n", err)
		return nil, err
	}
	return clientSet, err
}

// KubeConfig is set by the order of "inCluster" "~/.kube/config" and the env "KUBECONFIG"
func newConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		var kubeConfig string
		// fallback to kubeConfig
		if envHome := os.Getenv("HOME"); len(envHome) > 0 {
			kubeConfig = filepath.Join(envHome, ".kube", "config")
		}
		if envVar := os.Getenv("KUBECONFIG"); len(envVar) > 0 {
			kubeConfig = envVar
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			klog.Errorf("The kubeConfig cannot be loaded: %v\n", err)
			return nil, err
		}
	}
	return config, err
}
