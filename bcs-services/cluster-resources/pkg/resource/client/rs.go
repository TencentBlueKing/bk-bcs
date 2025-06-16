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
 */

package client

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/polymorphichelpers"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/action"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// RSClient ReplicaSet Client
type RSClient struct {
	ResClient
}

// NewRSClient ...
func NewRSClient(ctx context.Context, conf *res.ClusterConf) *RSClient {
	rsRes, _ := res.GetGroupVersionResource(ctx, conf, resCsts.RS, "")
	return &RSClient{ResClient{NewDynamicClient(conf), conf, rsRes}}
}

// NewRSCliByClusterID ...
func NewRSCliByClusterID(ctx context.Context, clusterID string) *RSClient {
	return NewRSClient(ctx, res.NewClusterConf(clusterID))
}

// List 获取 ReplicaSet 列表
func (c *RSClient) List(
	ctx context.Context, namespace, ownerName string, opts metav1.ListOptions,
) (map[string]interface{}, error) {
	ret, err := c.ResClient.List(ctx, namespace, opts)
	if err != nil {
		return nil, err
	}
	manifest := ret.UnstructuredContent()
	// 只有指定 OwnerReferences 信息才会再过滤
	if ownerName == "" {
		return manifest, nil
	}

	ownerRefs := []map[string]string{{"kind": resCsts.Deploy, "name": ownerName}}
	manifest["items"] = filterByOwnerRefs(mapx.GetList(manifest, "items"), ownerRefs)
	return manifest, nil
}

// GetResHistoryRevision 获取 workload history revision
func (c *RSClient) GetResHistoryRevision(ctx context.Context, kind, namespace, name string) ([]map[string]interface{},
	error) {

	// permValidate IAM 权限校验
	if err := c.permValidate(ctx, action.List, namespace); err != nil {
		return nil, err
	}

	// 初始化
	m := make([]map[string]interface{}, 0)

	clientSet, err := kubernetes.NewForConfig(c.conf.Rest)
	if err != nil {
		return m, err
	}

	// 通过Group创建HistoryViewer
	historyViewer, err := HistoryViewerFor(
		schema.GroupKind{Group: c.res.Group, Kind: kind}, clientSet)
	if err != nil {
		return m, err
	}

	// 获取 history
	s, err := historyViewer.GetHistory(namespace, name)
	if err != nil {
		return m, err
	}

	var versions []int64
	for k := range s {
		versions = append(versions, k)
	}
	SortInts64Desc(versions)

	for _, v := range versions {
		var unstructuredObj map[string]interface{}
		unstructuredObj, err = runtime.DefaultUnstructuredConverter.ToUnstructured(s[v])
		if err != nil {
			log.Error(ctx, "convert to unstructured failed, err %s", err.Error())
			continue
		}
		ret := formatter.FormatControllerRevisionRes(unstructuredObj)
		ret["revision"] = v
		m = append(m, ret)
	}
	return m, err
}

// GetResRevisionDiff 获取 workload revision差异信息
func (c *RSClient) GetResRevisionDiff(
	ctx context.Context, kind, namespace, name string, revision int64) (m map[string]interface{}, err error) {

	// permValidate IAM 权限校验
	if err = c.permValidate(ctx, action.View, namespace); err != nil {
		return nil, err
	}

	// 初始化
	m = map[string]interface{}{}

	// 初始化k8s ClientSet
	clientSet, err := kubernetes.NewForConfig(c.conf.Rest)
	if err != nil {
		return m, err
	}

	// 通过GroupKind创建HistoryViewer，获取template
	historyViewer, err := HistoryViewerFor(
		schema.GroupKind{Group: c.res.Group, Kind: kind}, clientSet)
	if err != nil {
		return m, err
	}

	// 以string的方法返回revision相关信息
	rolloutHistory, err := historyViewer.ViewHistory(namespace, name, revision)
	if err != nil {
		return m, err
	}

	currentHistory, err := historyViewer.ViewHistory(namespace, name, 0)
	if err != nil {
		return m, err
	}

	// key为revision，值为template，string格式
	m[resCsts.RolloutRevision] = rolloutHistory
	m[resCsts.CurrentRevision] = currentHistory
	return m, err
}

// RolloutResRevision 回滚某个资源 history revision
func (c *RSClient) RolloutResRevision(ctx context.Context, namespace, name, kind string, revision int64) error {

	// permValidate IAM 权限校验
	if err := c.permValidate(ctx, action.Update, namespace); err != nil {
		return err
	}

	clientSet, err := kubernetes.NewForConfig(c.conf.Rest)
	if err != nil {
		return err
	}
	appV1Cli := clientSet.AppsV1()
	// 根据kind获取对应资源客户端
	var deploy interface{}
	switch strings.ToLower(kind) {
	case "deployment":
		deploy, err = appV1Cli.Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
	case "statefulset":
		deploy, err = appV1Cli.StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
	case "daemonset":
		deploy, err = appV1Cli.DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s kind doesn't exist", kind)
	}
	rollBacker, err := polymorphichelpers.RollbackerFor(
		schema.GroupKind{Group: c.res.Group, Kind: kind}, clientSet)
	if err != nil {
		return err
	}
	object, ok := deploy.(runtime.Object)
	if !ok {
		return fmt.Errorf("%s Type assertion failed", kind)
	}
	_, err = rollBacker.Rollback(object, nil, revision, util.DryRunNone)
	return err
}
