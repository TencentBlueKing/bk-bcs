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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

// CRDClient ...
type CRDClient struct {
	ResClient
}

// NewCRDClient ...
func NewCRDClient(ctx context.Context, conf *res.ClusterConf) *CRDClient {
	CRDRes, _ := res.GetGroupVersionResource(ctx, conf, res.CRD, "")
	return &CRDClient{ResClient{NewDynamicClient(conf), conf, CRDRes}}
}

// NewCRDCliByClusterID ...
func NewCRDCliByClusterID(ctx context.Context, clusterID string) *CRDClient {
	return NewCRDClient(ctx, res.NewClusterConfig(clusterID))
}

// List ...
func (c *CRDClient) List(ctx context.Context, opts metav1.ListOptions) (map[string]interface{}, error) {
	ret, err := c.ResClient.List(ctx, "", opts)
	if err != nil {
		return nil, err
	}
	manifest := ret.UnstructuredContent()
	// 共享集群命名空间，需要过滤出属于指定项目的
	clusterInfo, err := cluster.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if clusterInfo.Type == cluster.ClusterTypeShared {
		crdList := []interface{}{}
		for _, crd := range manifest["items"].([]interface{}) {
			crdName := mapx.GetStr(crd.(map[string]interface{}), "metadata.name")
			if IsSharedClusterEnabledCRD(crdName) {
				crdList = append(crdList, crd)
			}
		}
		manifest["items"] = crdList
		return manifest, nil
	}
	return manifest, nil
}

// Watch ...
func (c *CRDClient) Watch(
	ctx context.Context, clusterType string, opts metav1.ListOptions,
) (watch.Interface, error) {
	rawWatch, err := c.ResClient.Watch(ctx, "", opts)
	return &CRDWatcher{rawWatch, clusterType}, err
}

// IsSharedClusterEnabledCRD 判断某 CRD，在共享集群中是否支持
func IsSharedClusterEnabledCRD(name string) bool {
	return slice.StringInSlice(name, conf.G.SharedCluster.EnabledCRDs)
}

// CRDWatcher ...
type CRDWatcher struct {
	watch.Interface

	clusterType string
}

// ResultChan ...
func (w *CRDWatcher) ResultChan() <-chan watch.Event {
	if w.clusterType == cluster.ClusterTypeSingle {
		return w.Interface.ResultChan()
	}
	// 共享集群，只能保留受支持的 CRD 的事件
	resultChan := make(chan watch.Event)
	go func() {
		for event := range w.Interface.ResultChan() {
			if obj, ok := event.Object.(*unstructured.Unstructured); ok {
				crdName := mapx.GetStr(obj.UnstructuredContent(), "metadata.name")
				if !IsSharedClusterEnabledCRD(crdName) {
					continue
				}
			}
			resultChan <- event
		}
	}()
	return resultChan
}

// GetCRDInfo 获取 CRD 基础信息
func GetCRDInfo(ctx context.Context, clusterID, crdName string) (map[string]interface{}, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	crdRes, err := res.GetGroupVersionResource(ctx, clusterConf, res.CRD, "")
	if err != nil {
		return nil, err
	}

	var ret *unstructured.Unstructured
	ret, err = NewResClient(clusterConf, crdRes).Get(ctx, "", crdName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	manifest := ret.UnstructuredContent()
	return formatter.FormatCRD(manifest), nil
}

// GetCObjManifest 获取自定义资源信息
func GetCObjManifest(
	ctx context.Context, clusterConf *res.ClusterConf, cobjRes schema.GroupVersionResource, namespace, cobjName string,
) (manifest map[string]interface{}, err error) {
	var ret *unstructured.Unstructured
	ret, err = NewResClient(clusterConf, cobjRes).Get(ctx, namespace, cobjName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return ret.UnstructuredContent(), nil
}
