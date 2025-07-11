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

// Package client xxx
package client

import (
	"context"
	"strings"

	"github.com/adevjoe/opentelemetry-go-contrib/instrumentation/k8s.io/client-go/transport"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/perm"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

var defaultLimit = 2000

// NewDynamicClient xxx
func NewDynamicClient(conf *res.ClusterConf) dynamic.Interface {
	conf.Rest.Wrap(transport.NewWrapperFunc())
	dynamicCli, _ := dynamic.NewForConfig(conf.Rest)
	return dynamicCli
}

// ResClient K8S 集群资源管理客户端
type ResClient struct {
	cli  dynamic.Interface
	conf *res.ClusterConf
	res  schema.GroupVersionResource
}

// NewResClient xxx
func NewResClient(conf *res.ClusterConf, resource schema.GroupVersionResource) *ResClient {
	return &ResClient{NewDynamicClient(conf), conf, resource}
}

// List 获取资源列表
func (c *ResClient) List(
	ctx context.Context, namespace string, opts metav1.ListOptions,
) (*unstructured.UnstructuredList, error) {
	if err := c.permValidate(ctx, action.List, namespace); err != nil {
		return nil, err
	}
	var object map[string]interface{}
	result := make([]unstructured.Unstructured, 0)
	opts.Limit = int64(defaultLimit)
	opts.Continue = ""
	for {
		ret, err := c.cli.Resource(c.res).Namespace(namespace).List(ctx, opts)
		if err != nil {
			return nil, c.handleErr(ctx, err)
		}
		object = ret.Object
		result = append(result, ret.Items...)
		if ret.GetContinue() == "" {
			break
		}
		opts.Continue = ret.GetContinue()
	}
	return &unstructured.UnstructuredList{Object: object, Items: result}, c.handleErr(ctx, nil)
}

// ListWithoutPerm 获取资源列表，不做权限校验
func (c *ResClient) ListWithoutPerm(
	ctx context.Context, namespace string, opts metav1.ListOptions,
) (*unstructured.UnstructuredList, error) {
	ret, err := c.cli.Resource(c.res).Namespace(namespace).List(ctx, opts)
	return ret, c.handleErr(ctx, err)
}

// ListAllWithoutPerm 获取全部资源列表，不做权限校验
func (c *ResClient) ListAllWithoutPerm(
	ctx context.Context, namespace string, opts metav1.ListOptions,
) ([]unstructured.Unstructured, error) {
	result := make([]unstructured.Unstructured, 0)
	opts.Limit = int64(defaultLimit)
	opts.Continue = ""
	for {
		ret, err := c.cli.Resource(c.res).Namespace(namespace).List(ctx, opts)
		if err != nil {
			return nil, c.handleErr(ctx, err)
		}
		result = append(result, ret.Items...)
		if ret.GetContinue() == "" {
			break
		}
		opts.Continue = ret.GetContinue()
	}
	return result, c.handleErr(ctx, nil)
}

// ListAllWithoutPermPreferred 获取全部资源列表，不做权限校验，没有数据不报错
func (c *ResClient) ListAllWithoutPermPreferred(
	ctx context.Context, namespace string, opts metav1.ListOptions) ([]unstructured.Unstructured, error) {
	result := make([]unstructured.Unstructured, 0)
	opts.Limit = int64(defaultLimit)
	opts.Continue = ""
	for {
		ret, err := c.cli.Resource(c.res).Namespace(namespace).List(ctx, opts)
		if err != nil {
			if errors.IsNotFound(err) {
				return []unstructured.Unstructured{}, nil
			}
			return nil, c.handleErr(ctx, err)
		}
		result = append(result, ret.Items...)
		if ret.GetContinue() == "" {
			break
		}
		opts.Continue = ret.GetContinue()
	}
	return result, c.handleErr(ctx, nil)
}

// Get 获取单个资源
func (c *ResClient) Get(
	ctx context.Context, namespace, name string, opts metav1.GetOptions,
) (*unstructured.Unstructured, error) {
	if err := c.permValidate(ctx, action.View, namespace); err != nil {
		return nil, err
	}
	ret, err := c.cli.Resource(c.res).Namespace(namespace).Get(ctx, name, opts)
	return ret, c.handleErr(ctx, err)
}

// GetWithoutPerm 获取单个资源
func (c *ResClient) GetWithoutPerm(
	ctx context.Context, namespace, name string, opts metav1.GetOptions,
) (*unstructured.Unstructured, error) {
	ret, err := c.cli.Resource(c.res).Namespace(namespace).Get(ctx, name, opts)
	return ret, c.handleErr(ctx, err)
}

// Create 创建资源
func (c *ResClient) Create(
	ctx context.Context, manifest map[string]interface{}, isNSScoped bool, opts metav1.CreateOptions,
) (*unstructured.Unstructured, error) {
	namespace := ""
	if isNSScoped {
		namespace = mapx.GetStr(manifest, "metadata.namespace")
		if namespace == "" {
			return nil, errorx.New(
				errcode.ValidateErr,
				i18n.GetMsg(ctx, "创建 %s 需要指定 metadata.namespace"),
				c.res.Resource,
			)
		}
	}
	if err := c.permValidate(ctx, action.Create, namespace); err != nil {
		return nil, err
	}
	ret, err := c.cli.Resource(c.res).Namespace(namespace).Create(
		ctx, &unstructured.Unstructured{Object: manifest}, opts)
	return ret, c.handleErr(ctx, err)
}

// Update 更新单个资源
func (c *ResClient) Update(
	ctx context.Context, namespace, name string, manifest map[string]interface{}, opts metav1.UpdateOptions,
) (*unstructured.Unstructured, error) {
	// 检查 name 与 manifest.metadata.name 是否一致
	manifestName, err := mapx.GetItems(manifest, "metadata.name")
	if err != nil || name != manifestName {
		return nil, errorx.New(
			errcode.ValidateErr,
			i18n.GetMsg(ctx, "metadata.name 必须指定且与准备编辑的资源名保持一致"),
		)
	}
	if err = c.permValidate(ctx, action.Update, namespace); err != nil {
		return nil, err
	}
	ret, err := c.cli.Resource(c.res).Namespace(namespace).Update(
		ctx, &unstructured.Unstructured{Object: manifest}, opts)
	return ret, c.handleErr(ctx, err)
}

// ApplyWithoutPerm 创建或更新资源，不做权限校验
func (c *ResClient) ApplyWithoutPerm(
	ctx context.Context, manifest map[string]interface{}, opts metav1.CreateOptions,
) (*unstructured.Unstructured, error) {
	name := mapx.GetStr(manifest, "metadata.name")
	namespace := mapx.GetStr(manifest, "metadata.namespace")
	if name == "" {
		return nil, errorx.New(errcode.ValidateErr, i18n.GetMsg(ctx, "metadata.name 必须指定"))
	}
	old, err := c.cli.Resource(c.res).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		if !errors.IsNotFound(err) {
			return nil, c.handleErr(ctx, err)
		}
		ret, errr := c.cli.Resource(c.res).Namespace(namespace).Create(
			ctx, &unstructured.Unstructured{Object: manifest}, opts)
		return ret, c.handleErr(ctx, errr)
	}
	_ = mapx.SetItems(manifest, "metadata.resourceVersion", old.GetResourceVersion())
	ret, err := c.cli.Resource(c.res).Namespace(namespace).Update(
		ctx, &unstructured.Unstructured{Object: manifest}, metav1.UpdateOptions{DryRun: opts.DryRun})
	return ret, c.handleErr(ctx, err)
}

// Patch 以 Patch 的方式更新资源
func (c *ResClient) Patch(
	ctx context.Context, namespace, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions,
) (*unstructured.Unstructured, error) {
	if err := c.permValidate(ctx, action.Update, namespace); err != nil {
		return nil, err
	}
	ret, err := c.cli.Resource(c.res).Namespace(namespace).Patch(ctx, name, pt, data, opts)
	return ret, c.handleErr(ctx, err)
}

// Delete 删除单个资源
func (c *ResClient) Delete(ctx context.Context, namespace, name string, opts metav1.DeleteOptions) error {
	if err := c.permValidate(ctx, action.Delete, namespace); err != nil {
		return err
	}
	// 若没有设置 PropagationPolicy，则设置为 Background
	// https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/#deleting-a-replicaset-and-its-pods
	if opts.PropagationPolicy == nil {
		policy := metav1.DeletePropagationBackground
		opts.PropagationPolicy = &policy
	}
	return c.handleErr(ctx, c.cli.Resource(c.res).Namespace(namespace).Delete(ctx, name, opts))
}

// DeleteWithoutPerm 删除单个资源
func (c *ResClient) DeleteWithoutPerm(ctx context.Context, namespace, name string, opts metav1.DeleteOptions) error {
	// 若没有设置 PropagationPolicy，则设置为 Background
	// https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/#deleting-a-replicaset-and-its-pods
	if opts.PropagationPolicy == nil {
		policy := metav1.DeletePropagationBackground
		opts.PropagationPolicy = &policy
	}
	return c.handleErr(ctx, c.cli.Resource(c.res).Namespace(namespace).Delete(ctx, name, opts))
}

// Watch 获取某类资源 watcher
func (c *ResClient) Watch(ctx context.Context, namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	if err := c.permValidate(ctx, action.List, namespace); err != nil {
		return nil, err
	}
	watcher, err := c.cli.Resource(c.res).Namespace(namespace).Watch(ctx, opts)
	return watcher, c.handleErr(ctx, err)
}

// permValidate IAM 权限校验
func (c *ResClient) permValidate(ctx context.Context, action, namespace string) error {
	projInfo, err := project.FromContext(ctx)
	if err != nil {
		return errorx.New(errcode.General, i18n.GetMsg(ctx, "由 Context 获取项目信息失败"))
	}
	clusterInfo, err := cluster.FromContext(ctx)
	if err != nil {
		return errorx.New(errcode.General, i18n.GetMsg(ctx, "由 Context 获取集群信息失败"))
	}
	return perm.Validate(ctx, c.res.Resource, action, projInfo.ID, clusterInfo.ID, namespace)
}

// 对一些特殊错误做处理，主要是集群升级导致资源缓存过期，做一次主动清理
func (c *ResClient) handleErr(ctx context.Context, originErr error) error {
	if originErr == nil {
		return nil
	}
	if !strings.Contains(originErr.Error(), "the server could not find the requested resource") {
		return originErr
	}
	cli, err := res.NewRedisCacheClient4Conf(ctx, c.conf)
	if err != nil {
		return originErr
	}
	// 检查缓存锁，如果近期已清理过缓存，可直接忽略
	_ = cli.ClearCache()
	return errorx.New(
		errcode.General,
		i18n.GetMsg(ctx, "检测到集群资源变动，尝试同步资源信息中。若近期有升级集群行为，请稍后重试，若依旧失败，请联系容器助手"),
	)
}
