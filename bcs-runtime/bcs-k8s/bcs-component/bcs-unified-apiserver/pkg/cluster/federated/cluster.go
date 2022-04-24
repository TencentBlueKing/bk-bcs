/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package federated

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/clientutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ClusterStor
type ClusterStor struct {
	members      []string
	masterClient *kubernetes.Clientset
	k8sClientMap map[string]*kubernetes.Clientset
}

// NewDeploymentStor
func NewClusterStor(masterClientId string, members []string) (*ClusterStor, error) {
	stor := &ClusterStor{members: members, k8sClientMap: make(map[string]*kubernetes.Clientset)}
	for _, k := range members {
		k8sClient, err := clientutil.GetClusternetClientByClusterId(k)
		if err != nil {
			return nil, err
		}
		stor.k8sClientMap[k] = k8sClient
	}
	masterClient, err := clientutil.GetClusternetClientByClusterId(masterClientId)
	if err != nil {
		return nil, err
	}
	stor.masterClient = masterClient
	return stor, nil
}

// GetServerGroups /api 返回
func (s *ClusterStor) GetAPIVersions(ctx context.Context) (*metav1.APIVersions, error) {
	v := &metav1.APIVersions{}
	if err := s.masterClient.RESTClient().Get().AbsPath(s.masterClient.LegacyPrefix).Do(ctx).Into(v); err != nil {
		return nil, err
	}
	return v, nil
}

// GetServerGroups /apis/v1 返回
func (s *ClusterStor) ServerCoreV1Resources(ctx context.Context) (*metav1.APIResourceList, error) {
	v := &metav1.APIResourceList{}
	if err := s.masterClient.RESTClient().Get().AbsPath(s.masterClient.LegacyPrefix + "/v1").Do(ctx).Into(v); err != nil {
		return nil, err
	}
	return v, nil
}

// GetServerGroups /apis 返回
func (s *ClusterStor) GetServerGroups(ctx context.Context) (*metav1.APIGroupList, error) {
	result, err := s.masterClient.ServerGroups()
	if err != nil {
		return nil, err
	}
	filtedGroups := []metav1.APIGroup{}
	for _, group := range result.Groups {
		// deployment, statefulsets apis
		if group.Name == "apps" {
			filtedGroups = append(filtedGroups, group)
		}

		// 可以使用 kubectl get APIService 命令
		if group.Name == "apiregistration.k8s.io" {
			filtedGroups = append(filtedGroups, group)
		}
	}
	result.Groups = filtedGroups
	return result, nil
}

// GetServerGroups /apis/{group}/{version} 返回
func (s *ClusterStor) ServerResourcesForGroupVersion(ctx context.Context, groupVersion string) (*metav1.APIResourceList, error) {
	return s.masterClient.ServerResourcesForGroupVersion("/apps/v1")
}
