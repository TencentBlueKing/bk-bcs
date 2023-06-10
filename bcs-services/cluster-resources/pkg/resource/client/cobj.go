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
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

// CRDClient xxx
type CRDClient struct {
	ResClient
}

// NewCRDClient xxx
func NewCRDClient(ctx context.Context, conf *res.ClusterConf) *CRDClient {
	CRDRes, _ := res.GetGroupVersionResource(ctx, conf, resCsts.CRD, "")
	return &CRDClient{ResClient{NewDynamicClient(conf), conf, CRDRes}}
}

// NewCRDCliByClusterID xxx
func NewCRDCliByClusterID(ctx context.Context, clusterID string) *CRDClient {
	return NewCRDClient(ctx, res.NewClusterConf(clusterID))
}

// List xxx
func (c *CRDClient) List(ctx context.Context, opts metav1.ListOptions) (map[string]interface{}, error) {
	// 共享集群 CRD 不做权限检查，直接过滤出允许的数类
	clusterInfo, err := cluster.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if clusterInfo.Type == cluster.ClusterTypeShared {
		var ret *unstructured.UnstructuredList
		ret, err = c.ResClient.cli.Resource(c.res).List(ctx, opts)
		if err != nil {
			return nil, err
		}
		manifest := ret.UnstructuredContent()
		crdList := []interface{}{}
		for _, crd := range mapx.GetList(manifest, "items") {
			crdName := mapx.GetStr(crd.(map[string]interface{}), "metadata.name")
			if IsSharedClusterEnabledCRD(crdName) {
				crdList = append(crdList, crd)
			}
		}
		manifest["items"] = crdList
		return manifest, nil
	}
	// 普通集群的 CRD，按集群域资源检查权限
	ret, err := c.ResClient.List(ctx, "", opts)
	if err != nil {
		return nil, err
	}
	return ret.UnstructuredContent(), nil
}

// Get xxx
func (c *CRDClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (map[string]interface{}, error) {
	// 共享集群 CRD 获取，如果在允许的数类内，不做权限检查
	clusterInfo, err := cluster.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if clusterInfo.Type == cluster.ClusterTypeShared {
		if !IsSharedClusterEnabledCRD(name) {
			return nil, errorx.New(errcode.Unsupported, i18n.GetMsg(ctx, "共享集群中不支持查看 CRD %s 信息"), name)
		}

		var ret *unstructured.Unstructured
		ret, err = c.ResClient.cli.Resource(c.res).Get(ctx, name, opts)
		if err != nil {
			return nil, err
		}
		return ret.UnstructuredContent(), nil
	}
	// 普通集群的 CRD，按集群域资源检查权限
	ret, err := c.ResClient.Get(ctx, "", name, opts)
	if err != nil {
		return nil, err
	}
	return ret.UnstructuredContent(), nil
}

// Watch xxx
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

// CRDWatcher xxx
type CRDWatcher struct {
	watch.Interface

	clusterType string
}

// ResultChan xxx
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
	manifest, err := NewCRDCliByClusterID(ctx, clusterID).Get(ctx, crdName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

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
