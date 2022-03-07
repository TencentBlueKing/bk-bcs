/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/project"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

const (
	// ProjCodeAnnoKey 项目 Code 在命名空间 Annotations 中的 Key
	ProjCodeAnnoKey = "io.tencent.bcs.projectcode"
)

// NSClient ...
type NSClient struct {
	ResClient
}

// NewNSClient ...
func NewNSClient(conf *res.ClusterConf) *NSClient {
	NSRes, _ := res.GetGroupVersionResource(conf, res.NS, "")
	return &NSClient{ResClient{NewDynamicClient(conf), conf, NSRes}}
}

// NewNSCliByClusterID ...
func NewNSCliByClusterID(clusterID string) *NSClient {
	return NewNSClient(res.NewClusterConfig(clusterID))
}

// List ...
func (c *NSClient) List(projectID string, opts metav1.ListOptions) (map[string]interface{}, error) {
	ret, err := c.ResClient.List("", opts)
	if err != nil {
		return nil, err
	}
	manifest := ret.UnstructuredContent()
	// 共享集群命名空间，需要过滤出属于指定项目的
	clusterInfo, err := cluster.GetClusterInfo(c.ResClient.conf.ClusterID)
	if err != nil {
		return nil, err
	}
	if clusterInfo.Type == cluster.ClusterTypeShared {
		projInfo, err := project.GetProjectInfo(projectID)
		if err != nil {
			return nil, err
		}
		projNSList := []interface{}{}
		for _, ns := range manifest["items"].([]interface{}) {
			if isProjNSinSharedCluster(ns.(map[string]interface{}), projInfo.Code) {
				projNSList = append(projNSList, ns)
			}
		}
		manifest["items"] = projNSList
		return manifest, nil
	}
	return manifest, nil
}

// Watch ...
func (c *NSClient) Watch(
	ctx context.Context, projectCode string, clusterType string, opts metav1.ListOptions,
) (watch.Interface, error) {
	rawWatch, err := c.ResClient.Watch(ctx, "", opts)
	return &NSWatcher{rawWatch, projectCode, clusterType}, err
}

// IsProjNSinSharedCluster 判断某命名空间，是否属于指定项目（仅共享集群有效）
func IsProjNSinSharedCluster(projectID, clusterID, namespace string) bool {
	if namespace == "" {
		return false
	}
	manifest, err := NewNSCliByClusterID(clusterID).Get("", namespace, metav1.GetOptions{})
	if err != nil {
		return false
	}
	projInfo, err := project.GetProjectInfo(projectID)
	if err != nil {
		return false
	}
	return isProjNSinSharedCluster(manifest.UnstructuredContent(), projInfo.Code)
}

// 判断某命名空间，是否属于指定项目（仅共享集群有效）
func isProjNSinSharedCluster(manifest map[string]interface{}, projectCode string) bool {
	// 规则：属于项目的命名空间满足以下两点，但这里只需要检查 annotations 即可
	//   1. 命名(name) 以 ieg-{project_code}- 开头
	//   2. annotations 中包含 io.tencent.bcs.projectcode: {project_code}
	return mapx.Get(manifest, []string{"metadata", "annotations", ProjCodeAnnoKey}, "") == projectCode
}

// NSWatcher ...
type NSWatcher struct {
	watch.Interface

	projectCode string
	clusterType string
}

// ResultChan ...
func (w *NSWatcher) ResultChan() <-chan watch.Event {
	if w.clusterType == cluster.ClusterTypeSingle {
		return w.Interface.ResultChan()
	}
	// 共享集群，只能保留项目拥有的命名空间的事件
	resultChan := make(chan watch.Event)
	go func() {
		for event := range w.Interface.ResultChan() {
			if obj, ok := event.Object.(*unstructured.Unstructured); ok {
				if !isProjNSinSharedCluster(obj.UnstructuredContent(), w.projectCode) {
					continue
				}
			}
			resultChan <- event
		}
	}()
	return resultChan
}
