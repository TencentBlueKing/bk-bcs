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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	gameAppsv1 "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/clientset/versioned"
	gameScheme "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/clientset/versioned/scheme"
	"github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"golang.org/x/sync/errgroup"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/describe"
	"k8s.io/kubectl/pkg/polymorphichelpers"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

const (
	rollbackSuccess = "rolled back"
	rollbackSkipped = "skipped rollback"
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

// ListCrdResources xxx
func ListCrdResources(
	ctx context.Context, clusterID string, opts metav1.ListOptions) (*v1.CustomResourceDefinitionList, error) {

	clusterConf := res.NewClusterConf(clusterID)
	crdClient, err := clientset.NewForConfig(clusterConf.Rest)
	if err != nil {
		return nil, err
	}

	// 获取 CRD 资源
	return crdClient.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
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

// GetWihtNoCheck xxx
func (c *CRDClient) GetWihtNoCheck(
	ctx context.Context, name string, opts metav1.GetOptions) (map[string]interface{}, error) {
	// 普通集群的 CRD，按集群域资源检查权限
	ret, err := c.ResClient.Get(ctx, "", name, opts)
	if err != nil {
		return nil, err
	}
	return ret.UnstructuredContent(), nil
}

// HistoryRevision 获取自定义资源的history revision
func (c *CRDClient) HistoryRevision(ctx context.Context, kind, namespace, name string) ([]map[string]interface{},
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

	gameClientSet, err := versioned.NewForConfig(c.conf.Rest)
	if err != nil {
		return m, err
	}

	// 通过Group创建HistoryViewer
	historyViewer, err := CustomHistoryViewerFor(
		schema.GroupKind{Group: c.res.Group, Kind: kind}, clientSet, gameClientSet)
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
func (c *CRDClient) GetResRevisionDiff(
	ctx context.Context, kind, namespace, name string, revision int64) (m map[string]interface{}, err error) {

	// permValidate IAM 权限校验
	if err = c.permValidate(ctx, action.View, namespace); err != nil {
		return nil, err
	}

	// 初始化
	m = map[string]interface{}{}

	clientSet, err := kubernetes.NewForConfig(c.conf.Rest)
	if err != nil {
		return m, err
	}

	gameClientSet, err := versioned.NewForConfig(c.conf.Rest)
	if err != nil {
		return m, err
	}

	// 通过Group创建HistoryViewer
	historyViewer, err := CustomHistoryViewerFor(
		schema.GroupKind{Group: c.res.Group, Kind: kind}, clientSet, gameClientSet)
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

// RolloutRevision 自定义资源回滚history revision
func (c *CRDClient) RolloutRevision(ctx context.Context, namespace, name, kind string, revision int64) error {
	// permValidate IAM 权限校验
	if err := c.permValidate(ctx, action.Update, namespace); err != nil {
		return err
	}
	clientSet, err := kubernetes.NewForConfig(c.conf.Rest)
	if err != nil {
		return err
	}
	gameClientSet, err := versioned.NewForConfig(c.conf.Rest)
	if err != nil {
		return err
	}
	gameCli := gameClientSet.TkexV1alpha1()
	// 根据kind获取对应资源客户端
	var deploy interface{}
	var rollBacker polymorphichelpers.Rollbacker
	switch strings.ToLower(kind) {
	case "gamedeployment":
		rollBacker = &GameDeploymentRollbacker{
			c: clientSet,
			g: gameCli,
		}
		deploy, err = gameCli.GameDeployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
	case "gamestatefulset":
		rollBacker = &GameStatefulSetRollbacker{
			c: clientSet,
			g: gameCli,
		}
		deploy, err = gameCli.GameStatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s kind doesn't exist", kind)
	}

	object, ok := deploy.(runtime.Object)
	if !ok {
		return fmt.Errorf("%s Type assertion failed", kind)
	}
	_, err = rollBacker.Rollback(object, nil, revision, util.DryRunNone)
	return err
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

// GetClustersCRDInfo 获取多集群 CRD 基础信息
func GetClustersCRDInfo(ctx context.Context, clusterIDs []string, crdName string) (map[string]interface{}, error) {
	errGroup := errgroup.Group{}
	mux := sync.Mutex{}
	var result map[string]interface{}
	for _, v := range clusterIDs {
		clusterID := v
		errGroup.Go(func() error {
			cluterInfo, err := cluster.GetClusterInfo(ctx, clusterID)
			if err != nil {
				return err
			}
			ctx = context.WithValue(ctx, ctxkey.ClusterKey, cluterInfo)
			manifest, err := NewCRDCliByClusterID(ctx, clusterID).Get(ctx, crdName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			mux.Lock()
			defer mux.Unlock()
			result = manifest
			return nil
		})
	}
	if err := errGroup.Wait(); err != nil && result == nil {
		return nil, err
	}

	return formatter.FormatCRD(result), nil
}

// GetClustersCRDInfoDirect 直接获取多集群 CRD 基础信息
func GetClustersCRDInfoDirect(
	ctx context.Context, clusterIDs []string, crdName string) (map[string]interface{}, error) {
	errGroup := errgroup.Group{}
	mux := sync.Mutex{}
	var result map[string]interface{}
	for _, v := range clusterIDs {
		clusterID := v
		errGroup.Go(func() error {
			cluterInfo, err := cluster.GetClusterInfo(ctx, clusterID)
			if err != nil {
				return err
			}
			ctx = context.WithValue(ctx, ctxkey.ClusterKey, cluterInfo)
			manifest, err := NewCRDCliByClusterID(ctx, clusterID).GetWihtNoCheck(ctx, crdName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			mux.Lock()
			defer mux.Unlock()
			result = manifest
			return nil
		})
	}
	if err := errGroup.Wait(); err != nil && result == nil {
		return nil, err
	}

	return formatter.FormatCRD(result), nil
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

// GameDeploymentRollbacker gameDeployment rollback
type GameDeploymentRollbacker struct {
	c kubernetes.Interface
	g v1alpha1.TkexV1alpha1Interface
}

// Rollback toRevision a non-negative integer, with 0 being reserved to indicate rolling back to previous configuration
func (r *GameDeploymentRollbacker) Rollback(obj runtime.Object, updatedAnnotations map[string]string, toRevision int64,
	dryRunStrategy util.DryRunStrategy) (string, error) {
	if toRevision < 0 {
		return "", fmt.Errorf("unable to find specified revision %v in history", r)
	}
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return "", fmt.Errorf("failed to create accessor for kind %v: %s", obj.GetObjectKind(), err.Error())
	}
	ds, history, err := gameDeploymentHistory(r.c.AppsV1(), r.g, accessor.GetNamespace(), accessor.GetName())
	if err != nil {
		return "", err
	}
	if toRevision == 0 && len(history) <= 1 {
		return "", fmt.Errorf("no last revision to roll back to")
	}

	toHistory := findHistory(toRevision, history)
	if toHistory == nil {
		return "", fmt.Errorf("unable to find specified revision %v in history", r)
	}

	if dryRunStrategy == util.DryRunClient {
		// nolint
		appliedSS, err := gDSApplyRevision(ds, toHistory)
		if err != nil {
			return "", err
		}
		return printPodTemplate(&appliedSS.Spec.Template)
	}

	// Skip if the revision already matches current StatefulSet
	done, err := gameDeploymentMatch(ds, toHistory)
	if err != nil {
		return "", err
	}
	if done {
		return fmt.Sprintf("%s (current template already matches revision %d)", rollbackSkipped, toRevision), nil
	}

	patchOptions := metav1.PatchOptions{}
	if dryRunStrategy == util.DryRunServer {
		patchOptions.DryRun = []string{metav1.DryRunAll}
	}
	// Restore revision
	if _, err = r.g.GameDeployments(ds.Namespace).Patch(context.TODO(), ds.Name, types.MergePatchType,
		toHistory.Data.Raw, patchOptions); err != nil {
		return "", fmt.Errorf("failed restoring revision %d: %v", toRevision, err)
	}

	return rollbackSuccess, nil
}

// GameStatefulSetRollbacker gameStatefulSet rollback
type GameStatefulSetRollbacker struct {
	c kubernetes.Interface
	g v1alpha1.TkexV1alpha1Interface
}

// Rollback toRevision a non-negative integer, with 0 being reserved to indicate rolling back to previous configuration
func (r *GameStatefulSetRollbacker) Rollback(obj runtime.Object, updatedAnnotations map[string]string, toRevision int64,
	dryRunStrategy util.DryRunStrategy) (string, error) {
	if toRevision < 0 {
		return "", fmt.Errorf("unable to find specified revision %v in history", toRevision)
	}
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return "", fmt.Errorf("failed to create accessor for kind %v: %s", obj.GetObjectKind(), err.Error())
	}
	sts, history, err := gameStatefulSetHistory(r.c.AppsV1(), r.g, accessor.GetNamespace(), accessor.GetName())
	if err != nil {
		return "", err
	}
	if toRevision == 0 && len(history) <= 1 {
		return "", fmt.Errorf("no last revision to roll back to")
	}

	toHistory := findHistory(toRevision, history)
	if toHistory == nil {
		return "", fmt.Errorf("unable to find specified revision %v in history", toRevision)
	}

	if dryRunStrategy == util.DryRunClient {
		// nolint
		appliedSS, err := gSTSApplyRevision(sts, toHistory)
		if err != nil {
			return "", err
		}
		return printPodTemplate(&appliedSS.Spec.Template)
	}

	// Skip if the revision already matches current StatefulSet
	done, err := gameStatefulSetMatch(sts, toHistory)
	if err != nil {
		return "", err
	}
	if done {
		return fmt.Sprintf("%s (current template already matches revision %d)", rollbackSkipped, toRevision), nil
	}

	patchOptions := metav1.PatchOptions{}
	if dryRunStrategy == util.DryRunServer {
		patchOptions.DryRun = []string{metav1.DryRunAll}
	}
	// Restore revision
	if _, err = r.g.GameStatefulSets(sts.Namespace).Patch(context.TODO(), sts.Name, types.MergePatchType,
		toHistory.Data.Raw, patchOptions); err != nil {
		return "", fmt.Errorf("failed restoring revision %d: %v", toRevision, err)
	}

	return rollbackSuccess, nil
}

var appsCodec = gameScheme.Codecs.LegacyCodec(gameAppsv1.GroupVersion)

// gDSApplyRevision returns a new GameDeployment constructed by restoring the state in revision to set.
// If the returned error is nil, the returned GameDeployment is valid.
func gDSApplyRevision(set *gameAppsv1.GameDeployment, revision *appsv1.ControllerRevision) (*gameAppsv1.GameDeployment,
	error) {
	patched, err := strategicpatch.StrategicMergePatch([]byte(runtime.EncodeOrDie(appsCodec, set)),
		revision.Data.Raw, set)
	if err != nil {
		return nil, err
	}
	result := &gameAppsv1.GameDeployment{}
	err = json.Unmarshal(patched, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// gSTSApplyRevision returns a new GameStatefulSet constructed by restoring the state in revision to set.
// If the returned error is nil, the returned GameStatefulSet is valid.
func gSTSApplyRevision(set *gameAppsv1.GameStatefulSet,
	revision *appsv1.ControllerRevision) (*gameAppsv1.GameStatefulSet, error) {
	patched, err := strategicpatch.StrategicMergePatch([]byte(runtime.EncodeOrDie(appsCodec, set)),
		revision.Data.Raw, set)
	if err != nil {
		return nil, err
	}
	result := &gameAppsv1.GameStatefulSet{}
	err = json.Unmarshal(patched, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// gameDeploymentMatch check if the given Deployment's template matches the template stored in the given history.
func gameDeploymentMatch(ss *gameAppsv1.GameDeployment, history *appsv1.ControllerRevision) (bool, error) {
	patch, err := getGameDeploymentPatch(ss)
	if err != nil {
		return false, err
	}
	return bytes.Equal(patch, history.Data.Raw), nil
}

// gameStatefulSetMatch check if the given StatefulSet's template matches the template stored in the given history.
func gameStatefulSetMatch(ss *gameAppsv1.GameStatefulSet, history *appsv1.ControllerRevision) (bool, error) {
	patch, err := getGameStatefulSetPatch(ss)
	if err != nil {
		return false, err
	}
	return bytes.Equal(patch, history.Data.Raw), nil
}

// getStatefulSetPatch returns a strategic merge patch that can be applied to restore a Deployment to a
// previous version. If the returned error is nil the patch is valid. The current state that we save is just the
// PodSpecTemplate. We can modify this later to encompass more state (or less) and remain compatible with previously
// recorded patches.
func getGameDeploymentPatch(set *gameAppsv1.GameDeployment) ([]byte, error) {
	str, err := runtime.Encode(appsCodec, set)
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if err = json.Unmarshal(str, &raw); err != nil {
		return nil, err
	}
	objCopy := make(map[string]interface{})
	specCopy := make(map[string]interface{})
	spec := raw["spec"].(map[string]interface{})
	template := spec["template"].(map[string]interface{})
	specCopy["template"] = template
	template["$patch"] = "replace"
	objCopy["spec"] = specCopy
	patch, err := json.Marshal(objCopy)
	return patch, err
}

// getGameStatefulSetPatch returns a strategic merge patch that can be applied to restore a StatefulSet to a
// previous version. If the returned error is nil the patch is valid. The current state that we save is just the
// PodSpecTemplate. We can modify this later to encompass more state (or less) and remain compatible with previously
// recorded patches.
func getGameStatefulSetPatch(set *gameAppsv1.GameStatefulSet) ([]byte, error) {
	str, err := runtime.Encode(appsCodec, set)
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if err = json.Unmarshal(str, &raw); err != nil {
		return nil, err
	}
	objCopy := make(map[string]interface{})
	specCopy := make(map[string]interface{})
	spec := raw["spec"].(map[string]interface{})
	template := spec["template"].(map[string]interface{})
	specCopy["template"] = template
	template["$patch"] = "replace"
	objCopy["spec"] = specCopy
	patch, err := json.Marshal(objCopy)
	return patch, err
}

// printPodTemplate converts a given pod template into a human-readable string.
func printPodTemplate(specTemplate *corev1.PodTemplateSpec) (string, error) {
	podSpec, err := printTemplate(specTemplate)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("will roll back to %s", podSpec), nil
}

// NOCC:golint/unparam(设计如此)
// nolint
func printTemplate(template *corev1.PodTemplateSpec) (string, error) {
	buf := bytes.NewBuffer([]byte{})
	w := describe.NewPrefixWriter(buf)
	describe.DescribePodTemplate(template, w)
	return buf.String(), nil
}

// 简单的返回特殊的字段
// nolint
func getApiResourcesManifest(ctx context.Context, crdName, clusterID string) (map[string]interface{}, error) {
	manifest := make(map[string]interface{}, 0)
	// 分离资源名称
	s := strings.SplitN(crdName, ".", 2)
	if len(s) > 0 {
		crdName = s[0]
	}

	// group可能为空
	group := ""
	if len(s) > 1 {
		group = s[1]
	}
	apiResources, err := res.GetApiResources(ctx, res.NewClusterConf(clusterID), "", crdName)
	if err != nil {
		return nil, err
	}
	// 避免不了存在resources,group一样，版本不一样的GroupVersionResource
	for _, v := range apiResources {
		for _, vv := range v {
			if vv["resource"] == crdName && vv["group"] == group {
				manifest = map[string]interface{}{
					"name":       vv["resource"],
					"kind":       vv["kind"],
					"apiVersion": fmt.Sprintf("%s/%s", vv["group"], vv["version"]),
					"scope":      constants.ClusterScope,
				}

				if vv["group"] == "" {
					manifest["apiVersion"] = fmt.Sprintf("%s", vv["version"])
				}
				if vv["namespaced"] == true {
					manifest["scope"] = constants.NamespacedScope
				}
			}
		}
	}
	return manifest, nil
}
