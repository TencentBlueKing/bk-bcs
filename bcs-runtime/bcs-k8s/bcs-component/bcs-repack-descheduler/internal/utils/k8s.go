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
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	gameversioned "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/clientset/versioned"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/client/clientset/versioned"
)

// SupportEviction uses Discovery API to find out if the server support eviction subresource
// If supported, it will return its groupVersion; Otherwise, it will return ""
func SupportEviction(client kubernetes.Interface) (string, error) {
	discoveryClient := client.Discovery()
	groupList, err := discoveryClient.ServerGroups()
	if err != nil {
		return "", errors.Wrapf(err, "get kubernetes server groups failed")
	}
	foundPolicyGroup := false
	var policyGroupVersion string
	for _, group := range groupList.Groups {
		if group.Name == "policy" {
			foundPolicyGroup = true
			policyGroupVersion = group.PreferredVersion.GroupVersion
			break
		}
	}
	if !foundPolicyGroup {
		return "", errors.Errorf("not found policy group")
	}
	resourceList, err := discoveryClient.ServerResourcesForGroupVersion("v1")
	if err != nil {
		return "", errors.Wrapf(err, "get v1 server resources failed")
	}
	for _, resource := range resourceList.APIResources {
		if resource.Name == apis.EvictionSubresource && resource.Kind == apis.EvictionKind {
			return policyGroupVersion, nil
		}
	}
	return "", nil
}

// SupportPDB check the kubernetes cluster support pdb
func SupportPDB(client kubernetes.Interface) (string, error) {
	discoveryClient := client.Discovery()
	_, err := discoveryClient.ServerResourcesForGroupVersion(apis.PDBGroupV1Version)
	if err == nil {
		return apis.PDBGroupV1Version, nil
	}
	_, err = discoveryClient.ServerResourcesForGroupVersion(apis.PDBGroupBetaVersion)
	if err == nil {
		return apis.PDBGroupBetaVersion, nil
	}
	return "", errors.Errorf("pdb not supported")
}

// SupportGameWorkload check the cluster support game workload
func SupportGameWorkload(client kubernetes.Interface, gwName, gwKind string) (bool, error) {
	discoveryClient := client.Discovery()
	resourceList, err := discoveryClient.ServerResourcesForGroupVersion(apis.BcsGroupVersion)
	if err != nil {
		return false, errors.Wrapf(err, "list resources for %s failed", apis.BcsGroupVersion)
	}
	for _, resource := range resourceList.APIResources {
		if resource.Name == gwName && resource.Kind == gwKind {
			return true, nil
		}
	}
	return false, nil
}

// GetK8sInClusterClient will return the in-cluster client of k8s.
func GetK8sInClusterClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "read in-cluster config failed")
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "create kubernetes client failed")
	}
	return client, nil
}

// GetK8sOutOfClusterClient return kubernetes client with kubeconfig path
func GetK8sOutOfClusterClient(kubeConfigPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, "build config failed")
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "new for config failed")
	}
	return clientSet, nil
}

// GetDeschedulePolicyClient return client for deschedulepolicy
func GetDeschedulePolicyClient() (*versioned.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "get in cluster config failed")
	}
	client, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "create client for deschedulepolicy failed")
	}
	return client, err
}

// GetGameDeploymentClient will return the in-cluster client of k8s.
func GetGameDeploymentClient() (*gameversioned.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "read in-cluster config failed")
	}
	client, err := gameversioned.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "create kubernetes client for gamedeployment failed")
	}
	return client, nil
}

// GetGameDeployClientWithKubeCfg get game deployment client
func GetGameDeployClientWithKubeCfg(kubeConfigPath string) (*gameversioned.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, "build config failed")
	}
	client, err := gameversioned.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "create kubernetes client for gamedeployment failed")
	}
	return client, nil
}

// GetGameStatefulSetClient will return the in-cluster client of k8s.
func GetGameStatefulSetClient() (*gameversioned.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "read in-cluster config failed")
	}
	client, err := gameversioned.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "create kubernetes client for gamestatefulset failed")
	}
	return client, nil
}

// GetGameStatefulSetCliWithKubeCfg get game statefulset client
func GetGameStatefulSetCliWithKubeCfg(kubeConfigPath string) (*gameversioned.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, "build config failed")
	}
	client, err := gameversioned.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "create kubernetes client for gamestatefulset failed")
	}
	return client, nil
}
