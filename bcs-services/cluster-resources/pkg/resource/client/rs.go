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
	"fmt"
	"strconv"
	"strings"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/polymorphichelpers"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/action"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
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

// GetDeployHistoryRevision 获取deployment history revision
func (c *RSClient) GetDeployHistoryRevision(
	ctx context.Context, deployName, namespace string) (m map[string]interface{}, err error) {

	// permValidate IAM 权限校验
	if err := c.permValidate(ctx, action.List, namespace); err != nil {
		return nil, err
	}

	// 初始化
	m = map[string]interface{}{}

	clientSet, err := kubernetes.NewForConfig(c.conf.Rest)
	if err != nil {
		return m, err
	}

	// 通过Group创建HistoryViewer
	historyViewer, err := polymorphichelpers.HistoryViewerFor(
		schema.GroupKind{Group: c.res.Group, Kind: "Deployment"}, clientSet)
	if err != nil {
		return m, err
	}

	// 获取deploy history
	s, err := historyViewer.GetHistory(namespace, deployName)
	if err != nil {
		return m, err
	}

	// 获取版本号和change cause map[int64]runtime.Object -> map[string]interface{}
	for key, data := range s {
		if value, ok := data.(*v1.ReplicaSet); ok {
			m[fmt.Sprintf("%d", key)] = value.ObjectMeta.Annotations[resCsts.ChangeCause]
		}
	}

	return m, err
}

// GetDeployRevisionDiff 获取deployment revision差异信息
func (c *RSClient) GetDeployRevisionDiff(
	ctx context.Context, deployName, namespace, revision string) (m map[string]interface{}, err error) {

	// permValidate IAM 权限校验
	if err := c.permValidate(ctx, action.View, namespace); err != nil {
		return nil, err
	}

	// 初始化
	m = map[string]interface{}{}

	// 即将回滚的版本，转换成int64，和前端的交互统一为string
	rolloutRevision, err := stringx.GetInt64(revision)
	if err != nil {
		return m, nil
	}

	// 初始化k8s ClientSet
	clientSet, err := kubernetes.NewForConfig(c.conf.Rest)
	if err != nil {
		return m, err
	}

	// 获取当前版本deploy相关信息
	deploy, err := clientSet.AppsV1().Deployments(namespace).Get(ctx, deployName, metav1.GetOptions{})
	if err != nil {
		return m, err
	}

	// 版本号
	revisionStr := deploy.Annotations[resCsts.Revision]
	// 转换成int64
	currentRevision, err := stringx.GetInt64(revisionStr)
	if err != nil {
		return m, nil
	}

	// 通过GroupKind创建HistoryViewer，获取template
	historyViewer, err := polymorphichelpers.HistoryViewerFor(
		schema.GroupKind{Group: c.res.Group, Kind: "Deployment"}, clientSet)
	if err != nil {
		return m, err
	}

	// 以string的方法返回revision相关信息
	rolloutHistory, err := historyViewer.ViewHistory(namespace, deployName, rolloutRevision)
	if err != nil {
		return m, err
	}

	currentHistory, err := historyViewer.ViewHistory(namespace, deployName, currentRevision)
	if err != nil {
		return m, err
	}

	// key为revision，值为template，string格式
	m[resCsts.RolloutRevision] = rolloutHistory
	m[resCsts.CurrentRevision] = currentHistory
	return m, err
}

// RolloutDeployRevision 回滚deployment history revision
func (c *RSClient) RolloutDeployRevision(
	ctx context.Context, namespace, revision, deployName string) (m map[string]interface{}, err error) {

	// permValidate IAM 权限校验
	if err := c.permValidate(ctx, action.Update, namespace); err != nil {
		return nil, err
	}

	// 初始化
	m = map[string]interface{}{}

	// 转换成int64，和前端的交互统一为string
	deployRevision, err := stringx.GetInt64(revision)
	if err != nil {
		return m, nil
	}

	clientSet, err := kubernetes.NewForConfig(c.conf.Rest)
	if err != nil {
		return m, err
	}

	// 获取deploy相关信息
	deploy, err := clientSet.AppsV1().Deployments(namespace).Get(ctx, deployName, metav1.GetOptions{})
	if err != nil {
		return m, err
	}

	// 通过deployment获取 rollbacker
	rollbacker, err := polymorphichelpers.RollbackerFor(
		schema.GroupKind{Group: c.res.Group, Kind: "Deployment"}, clientSet)
	if err != nil {
		return m, err
	}

	// rollout 回滚
	_, err = rollbacker.Rollback(deploy, nil, deployRevision, 0)
	if err != nil {
		return m, err
	}

	m["status"] = "ok"
	return m, err
}

// GetResHistoryRevision 获取某个资源 history revision
func (c *RSClient) GetResHistoryRevision(
	ctx context.Context, deployName, namespace, kind, changeCause string) (m map[string]interface{}, err error) {

	// permValidate IAM 权限校验
	if err := c.permValidate(ctx, action.List, namespace); err != nil {
		return nil, err
	}

	// 初始化
	m = map[string]interface{}{}

	clientSet, err := kubernetes.NewForConfig(c.conf.Rest)
	if err != nil {
		return m, err
	}

	historyViewer, err := polymorphichelpers.HistoryViewerFor(
		schema.GroupKind{Group: c.res.Group, Kind: kind}, clientSet)
	if err != nil {
		return m, err
	}

	s, err := historyViewer.GetHistory(namespace, deployName)
	if err != nil {
		return m, err
	}

	// 获取版本号和change cause map[int64]runtime.Object -> map[string]interface{}
	for key, data := range s {
		if value, ok := data.(metav1.Object); ok {
			m[strconv.FormatInt(key, 10)] = value.GetAnnotations()[changeCause]
		}
	}

	return m, err
}

// RolloutResRevision 回滚某个资源 history revision
func (c *RSClient) RolloutResRevision(
	ctx context.Context, namespace, revision, name, kind string) error {

	// permValidate IAM 权限校验
	if err := c.permValidate(ctx, action.Update, namespace); err != nil {
		return err
	}

	// 转换成int64，和前端的交互统一为string
	deployRevision, err := stringx.GetInt64(revision)
	if err != nil {
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
	object, ok := deploy.(runtime.Object)
	if !ok {
		return fmt.Errorf("%s Type assertion failed", kind)
	}
	_, err = rollBacker.Rollback(object, nil, deployRevision, 0)
	return err
}
