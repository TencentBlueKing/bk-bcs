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

// Package handler ns service
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	clusterapi "github.com/clusternet/clusternet/pkg/apis/clusters/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	third "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/thirdparty"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	fedtasks "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/tasks"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/types"
	trd "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/pkg/bcsapi/thirdparty-service"
	federationmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/proto/bcs-federation-manager"
)

// GetFederationClusterNamespace 根据集群ID，ns名称获取联邦集群ns
func (f *FederationManager) GetFederationClusterNamespace(ctx context.Context,
	req *federationmgr.GetFederationClusterNamespaceRequest,
	resp *federationmgr.GetFederationClusterNamespaceResponse) error {

	blog.Infof("Received BcsFederationManager.GetFederationClusterNamespace request, req: %+v", req)

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate GetFederationClusterNamespace request failed, err: %s", err.Error()))
	}

	// 根据联邦集群代理ID获取 联邦集群ID
	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), req.ClusterId)
	if err != nil {
		return ErrReturn(resp,
			fmt.Sprintf("GetFederationClusterNamespace get federation cluster from federationmanager failed, "+
				"clusterId: %s, err: %s", req.ClusterId, err.Error()))
	}

	// 获取联邦集群ns
	ns, err := f.clusterCli.GetNamespace(fedCluster.HostClusterID, req.Namespace)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("GetNamespace failed, err: %s", err.Error()))
	}

	// check is federated namespace
	if !IsFederationNamespace(ns) {
		return ErrReturn(resp, fmt.Sprintf("GetNamespace failed, %s is not a federated namespace", ns.Name))
	}
	// marshal data
	var data *federationmgr.FederationClusterNamespaceData
	jsonData, err := json.MarshalIndent(ns, "", "  ")
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("GetNamespace.Marshal failed, err: %s", err.Error()))
	}

	originK8SData := string(jsonData)
	createTime := ns.CreationTimestamp.Time.Format(time.RFC3339)
	data = &federationmgr.FederationClusterNamespaceData{
		ClusterId:       fedCluster.FederationClusterID,
		Namespace:       ns.Name,
		Annotations:     ns.Annotations,
		ClusterAffinity: nil,
		State:           "",
		CreateTime:      createTime,
		UpdateTime:      "",
		OriginK8SData:   &originK8SData,
	}
	// build data
	clusterAffinity := new(federationmgr.NamespaceSubClusterAffinity)
	if val, ok := ns.Annotations[cluster.ClusterAffinityMode]; ok {
		clusterAffinity.Mode = val
		data.ClusterAffinity = clusterAffinity
	}
	// annotations
	if val, ok := ns.Annotations[cluster.ClusterAffinitySelector]; ok {
		var labelSelector *federationmgr.LabelSelector
		err = json.Unmarshal([]byte(val), &labelSelector)
		if err != nil {
			return ErrReturn(resp, fmt.Sprintf("GetNamespace.Unmarshal failed, err: %s", err.Error()))
		}
		clusterAffinity.LabelSelector = labelSelector
		data.ClusterAffinity = clusterAffinity
	}
	// time
	if val, ok := ns.Annotations[cluster.NamespaceUpdateTimestamp]; ok {
		data.UpdateTime = val
	}
	// state
	if val, ok := ns.Annotations[cluster.HostClusterNamespaceStatus]; ok {
		data.State = val
	}
	// return success
	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = data
	return nil
}

// ListFederationClusterNamespace list ns
func (f *FederationManager) ListFederationClusterNamespace(ctx context.Context,
	req *federationmgr.ListFederationClusterNamespaceRequest,
	resp *federationmgr.ListFederationClusterNamespaceResponse) error {

	blog.Infof("Received BcsFederationManager.ListFederationClusterNamespace request, req: %+v", req)

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate ListFederationClusterNamespace request failed, err: %s", err.Error()))
	}

	// 根据联邦集群代理ID获取 联邦集群ID
	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), req.ClusterId)
	if err != nil {
		return ErrReturn(resp,
			fmt.Sprintf("ListFederationClusterNamespace get federation cluster from federationmanager failed, "+
				"clusterId: %s, err: %s", req.ClusterId, err.Error()))
	}
	// list all namespace
	namespaces, err := f.clusterCli.ListNamespace(fedCluster.HostClusterID)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("ListNamespace failed, err: %s", err.Error()))
	}

	var dataList []*federationmgr.FederationClusterNamespaceData
	for _, ns := range namespaces {
		// check is federated namespace
		if !IsFederationNamespace(&ns) {
			continue
		}
		// unmarshal response
		jsonData, err := json.MarshalIndent(ns, "", "  ")
		if err != nil {
			return ErrReturn(resp, fmt.Sprintf("ListNamespace.Marshal failed, err: %s", err.Error()))
		}
		// format time
		createTime := ns.CreationTimestamp.Time.Format(time.RFC3339)
		str := string(jsonData)
		data := &federationmgr.FederationClusterNamespaceData{
			ClusterId:       fedCluster.FederationClusterID,
			Namespace:       ns.Name,
			Annotations:     ns.Annotations,
			ClusterAffinity: nil,
			State:           "",
			CreateTime:      createTime,
			UpdateTime:      "",
			OriginK8SData:   &str,
		}

		// clusterAffinity
		clusterAffinity := new(federationmgr.NamespaceSubClusterAffinity)
		if val, ok := ns.Annotations[cluster.ClusterAffinityMode]; ok {
			clusterAffinity.Mode = val
			data.ClusterAffinity = clusterAffinity
		}

		// labelSelector
		if val, ok := ns.Annotations[cluster.ClusterAffinitySelector]; ok {
			var labelSelector *federationmgr.LabelSelector
			err = json.Unmarshal([]byte(val), &labelSelector)
			if err != nil {
				return ErrReturn(resp, fmt.Sprintf("ListNamespace.Unmarshal failed, err: %s", err.Error()))
			}
			clusterAffinity.LabelSelector = labelSelector
			data.ClusterAffinity = clusterAffinity
		}
		// annotations
		if val, ok := ns.Annotations[cluster.NamespaceUpdateTimestamp]; ok {
			data.UpdateTime = val
		}
		// state
		if val, ok := ns.Annotations[cluster.HostClusterNamespaceStatus]; ok {
			data.State = val
		}

		dataList = append(dataList, data)
	}
	// return
	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = dataList
	return nil
}

/*
	1、检测&修改请求；
	2、创建联邦命名空间；
	3、创建quota；
	4、创建子集群命名空间 (
		4.1、构建异步任务，记录异步任务ID到联邦命名空间的annotations中，
		4.2、在异步任务中执行检查、创建quota、创建子集群命名空间...等等steps
		4.3、在最后一个step根据任务状态同步到联邦命名空间的annotataions中
	)
*/

// CreateFederationClusterNamespace 根据入参判断创建原生k8s、太极平台、算力平台的 namespace
// NOCC:CCN_threshold(工具误报:)
func (f *FederationManager) CreateFederationClusterNamespace(ctx context.Context,
	req *federationmgr.CreateFederationClusterNamespaceRequest,
	resp *federationmgr.CreateFederationClusterNamespaceResponse) error {

	blog.Infof("Received BcsFederationManager.CreateFederationClusterNamespace request, req: %+v", req)

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("CreateFederationClusterNamespace request failed, err: %s", err.Error()))
	}

	// 根据联邦集群代理ID获取 联邦集群ID
	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), req.ClusterId)
	if err != nil {
		return ErrReturn(resp,
			fmt.Sprintf("CreateFederationClusterNamespace get federation cluster from federationmanager failed, "+
				"clusterId: %s, err: %s", req.ClusterId, err.Error()))
	}
	// 检查是否已存在namespace
	err = f.checkNamespaceIsExist(fedCluster, req)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("checkNamespaceIsExist failed, err: %s", err.Error()))
	}
	// 参数校验后创建ns
	subClusterIds, err := f.checkAndCreateNamespace(ctx, req, fedCluster)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("checkAndCreateNamespace failed, err: %s", err.Error()))
	}

	// 获取异步任务的参数
	reqMap, err := f.transferCreateSubClusterNamespaceParam(fedCluster.HostClusterID, subClusterIds, req)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("CreateFederationClusterNamespace transfor failed err: %s", err.Error()))
	}
	status := cluster.NamespaceCreating
	// init task
	if len(reqMap) == 0 {
		status = cluster.NamespaceSuccess
		blog.Errorf("CreateFederationClusterNamespace taskParams is empty, HostClusterID %s, subClusterIds %+v, "+
			"req %+v", fedCluster.HostClusterID, subClusterIds, req)
	} else {
		reqParameterBytes, err := json.Marshal(reqMap)
		if err != nil {
			return ErrReturn(resp, fmt.Sprintf("CreateFederationClusterNamespace.Marshal failed,err: %s", err.Error()))
		}
		// build task
		blog.Infof("CreateFederationClusterNamespace task reqParameterBytes: %s", string(reqParameterBytes))
		t, err := fedtasks.NewHandleNamespaceQuotaTask(
			&fedtasks.HandleNamespaceQuotaOptions{
				HandleType:    cluster.CreateKey,
				FedClusterId:  fedCluster.FederationClusterID,
				HostClusterId: fedCluster.HostClusterID,
				Namespace:     req.Namespace,
				Parameter:     string(reqParameterBytes),
			}).BuildTask(req.Creator)
		if err != nil {
			return ErrReturn(resp, fmt.Sprintf("BuildTask error when create federation cluster namespace, err: %s", err.Error()))
		}

		blog.Infof("CreateFederationClusterNamespace.CreateMultiClusterResourceQuota task: %+v", t)
		federationClusterNamespace, err := f.clusterCli.GetNamespace(fedCluster.HostClusterID, req.Namespace)
		if err != nil {
			return ErrReturn(resp, fmt.Sprintf("CreateFederationClusterNamespace.GetNamespace failed,err: %s", err.Error()))
		}
		// 将 任务id，状态 写入到annotations中
		federationClusterNamespace.Annotations[cluster.CreateNamespaceTaskId] = t.GetTaskID()
		federationClusterNamespace.Annotations[cluster.HostClusterNamespaceStatus] = status
		federationClusterNamespace.Annotations[cluster.NamespaceUpdateTimestamp] = time.Now().Format(time.RFC3339)
		err = f.clusterCli.UpdateNamespace(fedCluster.HostClusterID, federationClusterNamespace)
		if err != nil {
			return ErrReturn(resp, fmt.Sprintf("CreateFederationClusterNamespace.UpdateNamespace failed,err: %s",
				err.Error()))
		}
		// run task
		if err = f.taskmanager.Dispatch(t); err != nil {
			return ErrReturn(resp, fmt.Sprintf("CreateFederationClusterNamespace run task failed,err: %s", err.Error()))
		}
	}
	// return success
	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	return nil
}

// checkAndCreateNamespace checked and create ns
func (f *FederationManager) checkAndCreateNamespace(ctx context.Context, req *federationmgr.CreateFederationClusterNamespaceRequest,
	fedCluster *store.FederationCluster) ([]string, error) {

	// 获取所有的subCluster
	listSubClusters, err := store.GetStoreModel().ListSubClusters(ctx, &store.SubClusterListOptions{
		FederationClusterID: fedCluster.FederationClusterID})
	if err != nil {
		return nil, err
	}

	subClusterIds := make([]string, 0)
	subClusterIdMap := make(map[string]string)
	// 创建ns时，入参有没有cluster-range,都需要判断联邦的所有子集群中是否有同名ns
	if len(listSubClusters) > 0 {
		for _, subCluster := range listSubClusters {
			subClusterIdMap[subCluster.SubClusterID] = subCluster.SubClusterName
		}

		// 获取入参的子集群范围
		if clusterRangeStr, ok := req.Annotations[cluster.FedNamespaceClusterRangeKey]; ok {
			if len(clusterRangeStr) != 0 {
				lower := strings.Split(clusterRangeStr, ",")
				for _, sc := range lower {
					subClusterId := strings.ToUpper(sc)
					subClusterIds = append(subClusterIds, subClusterId)
				}
			}
		}
		// 检查入参的子集群是否都存在
		for _, subClusterId := range subClusterIds {
			if subClusterIdMap[subClusterId] == "" {
				return nil, fmt.Errorf("subClusterId %s is not in federation", subClusterId)
			}
		}

		// 判断联邦的所有子集群中是否有同名ns
		for subClusterId, _ := range subClusterIdMap {
			err = f.checkCreateSubClusterNamespace(subClusterId, fedCluster.FederationClusterID, req.Namespace)
			if err != nil {
				return nil, err
			}
		}
	}

	// 创建联邦集群ns
	err = createFederationNamespace(fedCluster.HostClusterID, f, req)
	if err != nil {
		return nil, err
	}
	return subClusterIds, nil
}

// checkNamespaceIsExist check the ns is exist
func (f *FederationManager) checkNamespaceIsExist(fedCluster *store.FederationCluster,
	req *federationmgr.CreateFederationClusterNamespaceRequest) error {
	// 查询是否已存在联邦集群命名空间
	fedClusterNamespace, err := f.clusterCli.GetNamespace(fedCluster.HostClusterID, req.Namespace)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("get federation namespace failed, "+
			"clusterId: %s, namespace: %s, err: %s", req.ClusterId, req.Namespace, err.Error())
	}

	if fedClusterNamespace != nil {
		return fmt.Errorf("in fed cluster the namespace already exist, "+
			"clusterId: %s, namespace: %s", req.ClusterId, req.Namespace)
	}

	return nil
}

// checkCreateSubClusterNamespace 检测子集群是否有这个命名空间，避免联邦的命名空间在子集群（作为独立集群工作时）和联邦有冲突。
func (f *FederationManager) checkCreateSubClusterNamespace(subClusterId, federationClusterID, namespace string) error {
	subClusterNamespace, err := f.clusterCli.GetNamespace(subClusterId, namespace)
	if err != nil && !errors.IsNotFound(err) {
		blog.Errorf("checkCreateSubClusterNamespace get sub cluster namespace failed, "+
			"subClusterId: %s, err: %s", subClusterId, err.Error())
		return err
	}

	if subClusterNamespace != nil {
		blog.Errorf("createFederationNamespace failed the namespace already exist in subCluster "+
			"subClusterNamespace: %+v", subClusterNamespace)
		return fmt.Errorf("the namespace already exist")
	}

	return nil
}

// checkUpdateSubClusterNamespace 检测子集群是否有效
func (f *FederationManager) checkUpdateSubClusterNamespace(subClusterId, federationClusterID string) (*store.SubCluster, error) {
	subCluster, err := store.GetStoreModel().GetSubCluster(context.Background(), federationClusterID, subClusterId)
	if err != nil {
		blog.Errorf("checkSubCluster get sub cluster from federationmanager failed, "+
			"fedClusterId: %s, subClusterId: %s, err: %s", federationClusterID, subClusterId, err.Error())
		return nil, err
	}

	return subCluster, nil
}

// NOCC:CCN_threshold(工具误报:),tosa/fn_length(设计如此)
func (f *FederationManager) transferCreateSubClusterNamespaceParam(hostClusterId string, subClusterIds []string,
	req *federationmgr.CreateFederationClusterNamespaceRequest) (map[string]string, error) {

	reqMap := make(map[string]string)
	var forTaijiList []*trd.CreateNamespaceForTaijiV3Request
	var forSuanliList []*trd.CreateNamespaceForSuanliRequest
	var forNormalList []*types.HandleNormalNamespace
	var forHunbuList []*types.HandleHunBuNamespace
	if req.Annotations[cluster.FedNamespaceBkbcsProjectCodeKey] == cluster.FedNamespaceProjectCodeTest {
		req.Annotations[cluster.FedNamespaceProjectCodeKey] = cluster.FedNamespaceProjectCodeTest
	}

	// 遍历subClusterIds，获取子集群的ns
	for _, subClusterId := range subClusterIds {
		// 获取managedCluster 对象
		managedCluster, err := f.clusterCli.GetManagedCluster(hostClusterId, subClusterId)
		if err != nil {
			blog.Errorf("transferCreateSubClusterNamespaceParam.GetManagedCluster failed, subClusterId: %s, "+
				"err: %s", subClusterId, err.Error())
			return nil, err
		}

		if managedCluster == nil {
			blog.Errorf("ManagedCluster is empty, subClusterId: %s", subClusterId)
			return nil, fmt.Errorf("ManagedCluster is empty, subClusterId: %s", subClusterId)
		}
		// 判断managedCluster对象的labels的标识
		switch managedCluster.Labels[cluster.ManagedClusterTypeLabel] {
		case cluster.SubClusterForTaiji:
			forTaijiList = f.initTjCreateNamespaceParams(req, forTaijiList)
		case cluster.SubClusterForSuanli:
			forSuanliList = f.initSlCreateNamespaceParams(req, forSuanliList)
		case cluster.SubClusterForHunbu:
			forHunbuList = f.initHbCreateNamespaceParams(hostClusterId, managedCluster, req, forHunbuList, subClusterId)
		default:
			forNormalList = append(forNormalList, &types.HandleNormalNamespace{
				FedClusterId: hostClusterId,
				SubClusterId: subClusterId,
				Namespace:    req.Namespace,
				Annotations:  req.Annotations,
			})
		}
	}

	// 处理太极ns
	if len(forTaijiList) > 0 {
		forTaijiListBytes, err := json.Marshal(forTaijiList)
		if err != nil {
			return nil, err
		}
		reqMap[cluster.SubClusterForTaiji] = string(forTaijiListBytes)
	}
	// 处理算力ns
	if len(forSuanliList) > 0 {
		forSuanliListBytes, err := json.Marshal(forSuanliList)
		if err != nil {
			return nil, err
		}
		reqMap[cluster.SubClusterForSuanli] = string(forSuanliListBytes)
	}

	// 处理混部ns
	if len(forHunbuList) > 0 {
		forHunbuListBytes, err := json.Marshal(forHunbuList)
		if err != nil {
			return nil, err
		}
		reqMap[cluster.SubClusterForHunbu] = string(forHunbuListBytes)
	}
	// 处理普通ns
	if len(forNormalList) > 0 {
		forNormalListBytes, err := json.Marshal(forNormalList)
		if err != nil {
			return nil, err
		}
		reqMap[cluster.SubClusterForNormal] = string(forNormalListBytes)
	}
	// 处理quota
	if len(req.QuotaList) > 0 {
		optBytes, err := json.Marshal(req.QuotaList)
		if err != nil {
			blog.Errorf("CreateFederationClusterNamespace.Marshal failed, err: %s", err.Error())
			return nil, err
		}
		reqMap[cluster.ClusterQuotaKey] = string(optBytes)
	}
	// return map
	return reqMap, nil
}

// initHbCreateNamespaceParams init params
func (f *FederationManager) initHbCreateNamespaceParams(hostClusterId string,
	managedCluster *clusterapi.ManagedCluster,
	req *federationmgr.CreateFederationClusterNamespaceRequest,
	forHunbuList []*types.HandleHunBuNamespace, subClusterId string) []*types.HandleHunBuNamespace {

	hbNsAnnotations := make(map[string]string)
	if managedCluster.Labels[cluster.LabelsMixerClusterKey] == cluster.ValueIsTrue {
		hbNsAnnotations[cluster.AnnotationMixerClusterMixerNamespaceKey] = cluster.ValueIsTrue
		// 判断是否有混部集群TKE网络方案
		if managedCluster.Labels[cluster.LabelsMixerClusterTkeNetworksKey] != "" {
			hbNsAnnotations[cluster.AnnotationMixerClusterNetworksKey] = managedCluster.Labels[cluster.LabelsMixerClusterTkeNetworksKey]
		}
		// 是否有project code
		if req.Annotations[cluster.FedNamespaceProjectCodeKey] != "" {
			hbNsAnnotations[cluster.FedNamespaceProjectCodeKey] = req.Annotations[cluster.FedNamespaceProjectCodeKey]
		}
		// 是否有优先级
		if managedCluster.Labels[cluster.LabelsMixerClusterPriorityKey] == cluster.ValueIsTrue {
			hbNsAnnotations[cluster.AnnotationMixerClusterPreemptionPolicyKey] = cluster.MixerClusterPreemptionPolicyValue
			hbNsAnnotations[cluster.AnnotationMixerClusterPreemptionClassKey] = cluster.MixerClusterPreemptionClassValue
			hbNsAnnotations[cluster.AnnotationMixerClusterPreemptionValueKey] = cluster.MixerClusterPreemptionValue
		}

		// 混部集群
		forHunbuList = append(forHunbuList, &types.HandleHunBuNamespace{
			HandleNormalNamespace: types.HandleNormalNamespace{
				FedClusterId: hostClusterId,
				SubClusterId: subClusterId,
				Namespace:    req.Namespace,
				Annotations:  hbNsAnnotations,
			},
			Labels: managedCluster.Labels,
		})
	}
	return forHunbuList
}

// initSlCreateNamespaceParams init params
func (f *FederationManager) initSlCreateNamespaceParams(req *federationmgr.CreateFederationClusterNamespaceRequest,
	forSuanliList []*trd.CreateNamespaceForSuanliRequest) []*trd.CreateNamespaceForSuanliRequest {
	// 构建 创建 suanli ns 的参数
	bkBizId := req.Annotations[cluster.FedNamespaceBkBizId]
	bkModuleId := req.Annotations[cluster.FedNamespaceBkModuleId]
	subQuotaInfos := make([]*trd.NamespaceQuotaForSuanli, 0)
	for _, quota := range req.QuotaList {
		// suanli quota
		if val, ok := quota.Annotations[cluster.AnnotationKeyInstalledPlatform]; ok {
			if val != cluster.SubClusterForSuanli {
				continue
			}

			quotaResources := make(map[string]string)
			for _, k8SResource := range quota.ResourceList {
				quotaResources[k8SResource.ResourceName] = k8SResource.ResourceQuantity
			}

			slAttributes := make(map[string]string)
			for k, v := range quota.Attributes {
				slAttributes[k] = v
			}

			quotaInfo := &trd.NamespaceQuotaForSuanli{
				Name:              quota.Name,
				SubQuotaLabels:    slAttributes,
				SubQuotaResources: quotaResources,
			}
			subQuotaInfos = append(subQuotaInfos, quotaInfo)
		}
	}

	forSuanliList = append(forSuanliList, &trd.CreateNamespaceForSuanliRequest{
		Namespace:     req.Namespace,
		SubQuotaInfos: subQuotaInfos,
		BkBizId:       bkBizId,
		BkModuleId:    bkModuleId,
	})
	return forSuanliList
}

// initTjCreateNamespaceParams init params
func (f *FederationManager) initTjCreateNamespaceParams(req *federationmgr.CreateFederationClusterNamespaceRequest,
	forTaijiList []*trd.CreateNamespaceForTaijiV3Request) []*trd.CreateNamespaceForTaijiV3Request {
	// 构建 创建 taiji ns 的参数
	bkBizId := req.Annotations[cluster.FedNamespaceBkBizId]
	bkModuleId := req.Annotations[cluster.FedNamespaceBkModuleId]
	subQuotaInfos := make([]*trd.NamespaceQuotaForTaiji, 0)
	location := ""
	for _, quota := range req.QuotaList {
		// taiji quota
		if val, ok := quota.Annotations[cluster.AnnotationSubClusterForTaiji]; ok {
			location = val
			quotaResources := make(map[string]string)
			for _, k8SResource := range quota.ResourceList {
				quotaResources[k8SResource.ResourceName] = k8SResource.ResourceQuantity
			}

			// 转换为taiji参数 GPUName
			tjAttributes := make(map[string]string)
			for k, v := range quota.Attributes {
				if k == cluster.TaskGpuTypeKey {
					tjAttributes[cluster.TaijiGPUNameKey] = v
					continue
				}
				tjAttributes[k] = v
			}

			quotaInfo := &trd.NamespaceQuotaForTaiji{
				Name:              quota.Name,
				SubQuotaLabels:    tjAttributes,
				SubQuotaResources: quotaResources,
				Location:          val,
			}
			subQuotaInfos = append(subQuotaInfos, quotaInfo)
		}
	}

	isPrivateResource := false
	if req.Annotations[cluster.AnnotationIsPrivateResourceKey] == "true" {
		isPrivateResource = true
	}
	// location
	if location == "" {
		location = req.Annotations[cluster.AnnotationSubClusterForTaiji]
	}

	scheduleAlgorithm := req.Annotations[cluster.AnnotationScheduleAlgorithmKey]
	forTaijiList = append(forTaijiList, &trd.CreateNamespaceForTaijiV3Request{
		Namespace:         req.Namespace,
		BkBizId:           bkBizId,
		BkModuleId:        bkModuleId,
		SubQuotaInfos:     subQuotaInfos,
		Location:          location,
		IsPrivateResource: &isPrivateResource,
		ScheduleAlgorithm: &scheduleAlgorithm,
	})
	return forTaijiList
}

// NOCC:CCN_threshold(工具误报:),tosa/fn_length(设计如此)
func (f *FederationManager) transferUpdateSubClusterNamespaceParam(hostClusterId string, subClusterIdMap map[string]string,
	annotations map[string]string, req *federationmgr.UpdateFederationClusterNamespaceRequest) (
	map[string]string, map[string]string, error) {

	reqMap := make(map[string]string)
	createNamespaceParamMap := make(map[string]string)
	var forNormalList []*types.HandleNormalNamespace
	var forHunbuList []*types.HandleHunBuNamespace
	if req.Annotations[cluster.FedNamespaceBkbcsProjectCodeKey] == cluster.FedNamespaceProjectCodeTest {
		req.Annotations[cluster.FedNamespaceProjectCodeKey] = cluster.FedNamespaceProjectCodeTest
	}

	for subClusterId, _ := range subClusterIdMap {
		// 获取managedCluster 对象
		managedCluster, err := f.clusterCli.GetManagedCluster(hostClusterId, subClusterId)
		if err != nil {
			blog.Errorf("GetManagedCluster failed, subClusterId: %s, err: %s", subClusterId, err.Error())
			return nil, nil, err
		}
		// 避免空指针
		if managedCluster == nil {
			blog.Errorf("ManagedCluster is empty, subClusterId: %s", subClusterId)
			return nil, nil, fmt.Errorf("ManagedCluster is empty, subClusterId: %s", subClusterId)
		}
		// 判断managedCluster对象的labels的标识
		switch managedCluster.Labels[cluster.ManagedClusterTypeLabel] {
		case cluster.SubClusterForTaiji:
			err = f.initTjUpsetNamespaceParams(hostClusterId, subClusterId, req, annotations, createNamespaceParamMap)
			if err != nil {
				return nil, nil, err
			}
		case cluster.SubClusterForSuanli:
			err = f.initSlUpsetNamespaceParams(hostClusterId, subClusterId, req, annotations, createNamespaceParamMap)
			if err != nil {
				return nil, nil, err
			}
		case cluster.SubClusterForHunbu:
			flag, err := f.getCreateNamespaceParamMap(hostClusterId, subClusterId, managedCluster, req, createNamespaceParamMap)
			if err != nil {
				blog.Errorf("getCreateNamespaceParamMap failed, subClusterId: %s, err: %s", subClusterId, err.Error())
				return nil, nil, err
			}
			if !flag {
				continue
			}
			forHunbuList = initHbCreateNamespaceParams(hostClusterId, subClusterId, req.Namespace, req.Annotations,
				managedCluster.Labels, forHunbuList)
		default:
			// 普通集群
			flag, err := f.getCreateNamespaceParamMap(hostClusterId, subClusterId, managedCluster, req, createNamespaceParamMap)
			if err != nil {
				blog.Errorf("getCreateNamespaceParamMap failed, subClusterId: %s, err: %s", subClusterId, err.Error())
				return nil, nil, err
			}
			if !flag {
				continue
			}
			forNormalList = append(forNormalList, &types.HandleNormalNamespace{
				FedClusterId: hostClusterId,
				SubClusterId: subClusterId,
				Namespace:    req.Namespace,
				Annotations:  req.Annotations,
			})
		}
	}
	// 混部集群
	if len(forHunbuList) > 0 {
		forHunbuListBytes, err := json.Marshal(forHunbuList)
		if err != nil {
			return nil, nil, err
		}
		reqMap[cluster.SubClusterForHunbu] = string(forHunbuListBytes)
	}
	// 普通集群
	if len(forNormalList) > 0 {
		forNormalListBytes, err := json.Marshal(forNormalList)
		if err != nil {
			return nil, nil, err
		}
		reqMap[cluster.SubClusterForNormal] = string(forNormalListBytes)
	}
	return reqMap, createNamespaceParamMap, nil
}

// initHbCreateNamespaceParams init params
func initHbCreateNamespaceParams(hostClusterId, subClusterId, namespace string, annotations, labels map[string]string,
	list []*types.HandleHunBuNamespace) []*types.HandleHunBuNamespace {

	hbNsAnnotations := make(map[string]string)
	if labels[cluster.LabelsMixerClusterKey] == cluster.ValueIsTrue {
		hbNsAnnotations[cluster.AnnotationMixerClusterMixerNamespaceKey] = cluster.ValueIsTrue
		// 判断是否有混部集群TKE网络方案
		if labels[cluster.LabelsMixerClusterTkeNetworksKey] != "" {
			hbNsAnnotations[cluster.AnnotationMixerClusterNetworksKey] = labels[cluster.LabelsMixerClusterTkeNetworksKey]
		}
		// 是否有project code
		if annotations[cluster.FedNamespaceProjectCodeKey] != "" {
			hbNsAnnotations[cluster.FedNamespaceProjectCodeKey] = annotations[cluster.FedNamespaceProjectCodeKey]
		}
		// 是否有优先级
		if labels[cluster.LabelsMixerClusterPriorityKey] == cluster.ValueIsTrue {
			hbNsAnnotations[cluster.AnnotationMixerClusterPreemptionPolicyKey] = cluster.MixerClusterPreemptionPolicyValue
			hbNsAnnotations[cluster.AnnotationMixerClusterPreemptionClassKey] = cluster.MixerClusterPreemptionClassValue
			hbNsAnnotations[cluster.AnnotationMixerClusterPreemptionValueKey] = cluster.MixerClusterPreemptionValue
		}

		// 混部集群
		list = append(list, &types.HandleHunBuNamespace{
			HandleNormalNamespace: types.HandleNormalNamespace{
				FedClusterId: hostClusterId,
				SubClusterId: subClusterId,
				Namespace:    namespace,
				Annotations:  hbNsAnnotations,
			},
			Labels: labels,
		})
	}

	return list
}

// initSlUpsetNamespaceParams init params
func (f *FederationManager) initSlUpsetNamespaceParams(hostClusterId, subClusterId string,
	req *federationmgr.UpdateFederationClusterNamespaceRequest, annotations, paramMap map[string]string) error {

	bkBizId, bkModuleId := getBkModuleParams(req, annotations)
	// 查询算力子集群是否已存在namespace
	resp, err := f.checkSlSubClusterNamespace(req.Namespace)
	if err != nil && !strings.Contains(err.Error(), "namespace not register") {
		return err
	}

	// 当namespace未注册时，去新增
	if resp != nil && strings.Contains(resp.Message, "namespace not register") {
		err = f.getCreateNamespaceParamForSuanli(hostClusterId, bkBizId, bkModuleId, req, paramMap)
		if err != nil {
			blog.Errorf("initSlCreateNamespaceParams failed, subClusterId: %s, err: %s", subClusterId, err.Error())
			return err
		}
	}

	return nil
}

// initTjUpsetNamespaceParams init params
func (f *FederationManager) initTjUpsetNamespaceParams(hostClusterId, subClusterId string,
	req *federationmgr.UpdateFederationClusterNamespaceRequest, annotations,
	createNamespaceParamMap map[string]string) error {

	bkBizId, bkModuleId := getBkModuleParams(req, annotations)
	isPrivateResourceFlag, scheduleAlgorithm, location := getParams(req, annotations)
	// 查询太极是否已存在namespace
	resp, err := f.checkTjSubClusterNamespace(req.Namespace)
	if err != nil {
		return err
	}

	// 当namespace未注册时，才去新增
	if resp != nil && resp.Error != nil && strings.Contains(resp.Error.Message, "namespace not register") {
		err = f.getCreateNamespaceParamForTaiji(hostClusterId, scheduleAlgorithm, location, bkBizId, bkModuleId,
			isPrivateResourceFlag, req, createNamespaceParamMap)
		if err != nil {
			blog.Errorf("initTjCreateNamespaceParams failed, subClusterId: %s, err: %s", subClusterId, err.Error())
			return err
		}
	}

	return nil
}

// checkTjSubClusterNamespace 检查太极子集群中是否已存在ns
func (f *FederationManager) checkTjSubClusterNamespace(namespace string) (*trd.GetKubeConfigForTaijiResponse, error) {
	// 查询third party api 获取suanli kubeConfig
	resp, err := third.GetThirdpartyClient().GetKubeConfigForTaiji(namespace)
	if err != nil {
		blog.Errorf("GetKubeConfigForTaiji failed namespace: %s, err: %s", namespace, err.Error())
		return nil, err
	}

	return resp, nil
}

// checkSlSubClusterNamespace 检查算力子集群中是否已存在ns
func (f *FederationManager) checkSlSubClusterNamespace(namespace string) (*trd.GetKubeConfigForSuanliResponse, error) {
	// 查询third party api 获取taiji kubeConfig
	resp, err := third.GetThirdpartyClient().GetKubeConfigForSuanli(namespace)
	if err != nil {
		blog.Errorf("GetKubeConfigForSuanli failed namespace: %s, err: %s", namespace, err.Error())
		return nil, err
	}

	return resp, nil
}

// getCreateNamespaceParamMap 更新namespace时，若子集群范围扩增，则需要构建新增ns的参数
func (f *FederationManager) getCreateNamespaceParamMap(hostClusterId, subClusterId string,
	managedCluster *clusterapi.ManagedCluster, req *federationmgr.UpdateFederationClusterNamespaceRequest,
	reqMap map[string]string) (bool, error) {

	// 先检测所有的子集群是否有这个命名空间
	subNs, err := f.clusterCli.GetNamespace(subClusterId, req.Namespace)
	if err != nil && !errors.IsNotFound(err) {
		blog.Errorf("GetNamespace failed, subClusterId: %s, err: %s", subClusterId, err.Error())
		return false, err
	}

	flag := false
	if subNs != nil {
		flag = true
		return flag, nil
	}

	var forNormalList []*types.HandleNormalNamespace
	var forHunbuList []*types.HandleHunBuNamespace

	// 判断managedCluster对象的labels的标识
	switch managedCluster.Labels[cluster.ManagedClusterTypeLabel] {
	case cluster.SubClusterForHunbu:
		hbNsAnnotations := make(map[string]string)
		if managedCluster.Labels[cluster.LabelsMixerClusterKey] == cluster.ValueIsTrue {
			hbNsAnnotations[cluster.AnnotationMixerClusterMixerNamespaceKey] = cluster.ValueIsTrue
			// 判断是否有混部集群TKE网络方案
			if managedCluster.Labels[cluster.LabelsMixerClusterTkeNetworksKey] != "" {
				hbNsAnnotations[cluster.AnnotationMixerClusterNetworksKey] =
					managedCluster.Labels[cluster.LabelsMixerClusterTkeNetworksKey]
			}
			// 是否有project code
			if req.Annotations[cluster.FedNamespaceProjectCodeKey] != "" {
				hbNsAnnotations[cluster.FedNamespaceProjectCodeKey] = req.Annotations[cluster.FedNamespaceProjectCodeKey]
			}
			// 是否有优先级
			if managedCluster.Labels[cluster.LabelsMixerClusterPriorityKey] == cluster.ValueIsTrue {
				hbNsAnnotations[cluster.AnnotationMixerClusterPreemptionPolicyKey] = cluster.MixerClusterPreemptionPolicyValue
				hbNsAnnotations[cluster.AnnotationMixerClusterPreemptionClassKey] = cluster.MixerClusterPreemptionClassValue
				hbNsAnnotations[cluster.AnnotationMixerClusterPreemptionValueKey] = cluster.MixerClusterPreemptionValue
			}

			forHunbuList = append(forHunbuList, &types.HandleHunBuNamespace{
				HandleNormalNamespace: types.HandleNormalNamespace{
					FedClusterId: hostClusterId,
					SubClusterId: subClusterId,
					Namespace:    req.Namespace,
					Annotations:  hbNsAnnotations,
				},
				Labels: managedCluster.Labels,
			})
		}
	default:
		forNormalList = append(forNormalList, &types.HandleNormalNamespace{
			FedClusterId: hostClusterId,
			SubClusterId: subClusterId,
			Namespace:    req.Namespace,
			Annotations:  req.Annotations,
		})
	}

	// 混部集群
	if len(forHunbuList) > 0 {
		forHunbuListBytes, err := json.Marshal(forHunbuList)
		if err != nil {
			return flag, err
		}
		reqMap[cluster.SubClusterForHunbu] = string(forHunbuListBytes)
	}

	// 普通集群
	if len(forNormalList) > 0 {
		forNormalListBytes, err := json.Marshal(forNormalList)
		if err != nil {
			return flag, err
		}
		reqMap[cluster.SubClusterForNormal] = string(forNormalListBytes)
	}

	return flag, nil
}

// getCreateNamespaceParamForTaiji init create ns params
func (f *FederationManager) getCreateNamespaceParamForTaiji(
	hostClusterId, scheduleAlgorithm, location, bkBizId, bkModuleId string, isPrivateResource bool,
	req *federationmgr.UpdateFederationClusterNamespaceRequest, reqMap map[string]string) error {

	// get quotas
	mcResourceQuotas, err := f.clusterCli.ListNamespaceQuota(hostClusterId, req.Namespace)
	if err != nil {
		blog.Errorf("ListNamespaceQuota failed, namespace: %s, err: %s",
			req.Namespace, err.Error())
		return err
	}

	quotas := make([]*trd.NamespaceQuotaForTaiji, 0)
	for _, mcResourceQuota := range mcResourceQuotas.Items {
		if val, ok := mcResourceQuota.Annotations[cluster.AnnotationSubClusterForTaiji]; ok {
			location = val
			quotaResources := make(map[string]string)
			if mcResourceQuota.Spec.TotalQuota.Hard != nil {
				for name, quantity := range mcResourceQuota.Spec.TotalQuota.Hard {
					quotaResources[string(name)] = quantity.String()
				}
			}

			quotaInfo := &trd.NamespaceQuotaForTaiji{
				Name:              mcResourceQuota.Name,
				SubQuotaLabels:    mcResourceQuota.Spec.TaskSelector,
				SubQuotaResources: quotaResources,
				Location:          val,
			}

			quotas = append(quotas, quotaInfo)
		}
	}

	forTaijiList := make([]*trd.CreateNamespaceForTaijiV3Request, 0)
	// 构建 创建 taiji ns 的参数
	forTaijiList = append(forTaijiList, &trd.CreateNamespaceForTaijiV3Request{
		Location:          location,
		Namespace:         req.Namespace,
		Creator:           req.Operator,
		ScheduleAlgorithm: &scheduleAlgorithm,
		GpuPpClusterNames: nil,
		SubQuotaInfos:     quotas,
		IsPrivateResource: &isPrivateResource,
		BkBizId:           bkBizId,
		BkModuleId:        bkModuleId,
	})

	forTaijiListBytes, err := json.Marshal(forTaijiList)
	if err != nil {
		return err
	}

	reqMap[cluster.SubClusterForTaiji] = string(forTaijiListBytes)

	return nil
}

// getCreateNamespaceParamForSuanli init create ns params
func (f *FederationManager) getCreateNamespaceParamForSuanli(hostClusterId, bkBizId, bkModuleId string,
	req *federationmgr.UpdateFederationClusterNamespaceRequest, reqMap map[string]string) error {

	// get quotas
	mcResourceQuotas, err := f.clusterCli.ListNamespaceQuota(hostClusterId, req.Namespace)
	if err != nil {
		blog.Errorf("ListNamespaceQuota failed, namespace: %s, err: %s", req.Namespace, err.Error())
		return err
	}
	// build quotas
	quotas := make([]*trd.NamespaceQuotaForSuanli, 0)
	for _, mcResourceQuota := range mcResourceQuotas.Items {
		if val, ok := mcResourceQuota.Annotations[cluster.AnnotationKeyInstalledPlatform]; ok {
			if val != cluster.SubClusterForSuanli {
				continue
			}
			quotaResources := make(map[string]string)
			if mcResourceQuota.Spec.TotalQuota.Hard != nil {
				for name, quantity := range mcResourceQuota.Spec.TotalQuota.Hard {
					quotaResources[string(name)] = quantity.String()
				}
			}

			quotaInfo := &trd.NamespaceQuotaForSuanli{
				Name:              mcResourceQuota.Name,
				SubQuotaLabels:    mcResourceQuota.Spec.TaskSelector,
				SubQuotaResources: quotaResources,
			}

			quotas = append(quotas, quotaInfo)
		}
	}

	forSuanliList := make([]*trd.CreateNamespaceForSuanliRequest, 0)
	forSuanliList = append(forSuanliList, &trd.CreateNamespaceForSuanliRequest{
		Namespace:     req.Namespace,
		Creator:       req.Operator,
		SubQuotaInfos: quotas,
		BkBizId:       bkBizId,
		BkModuleId:    bkModuleId,
	})
	forSuanliListBytes, err := json.Marshal(forSuanliList)
	if err != nil {
		return err
	}

	reqMap[cluster.SubClusterForSuanli] = string(forSuanliListBytes)

	return nil
}

// NOCC:tosa/fn_length(设计如此)
func (f *FederationManager) transferDeleteSubClusterNamespace(
	hostClusterId string, subClusterIds []string) (string, string, error) {

	var hbSubClusterIdArr []string
	var nmSubClusterIdArr []string
	hbSubClusterId := ""
	nmSubClusterId := ""
	for _, subClusterId := range subClusterIds {
		// 获取managedCluster 对象
		managedCluster, err := f.clusterCli.GetManagedCluster(hostClusterId, subClusterId)
		if err != nil {
			blog.Errorf("DeleteNamespace.GetManagedCluster 获取 ManagedCluster failed, subClusterId: %s, err: %s",
				subClusterId, err.Error())
			return "", "", err
		}
		// 避免空指针
		if managedCluster == nil {
			blog.Errorf("ManagedCluster is empty, subClusterId: %s", subClusterId)
			return "", "", fmt.Errorf("ManagedCluster is empty, subClusterId: %s", subClusterId)
		}

		// 判断managedCluster对象的labels的标识
		switch managedCluster.Labels[cluster.ManagedClusterTypeLabel] {
		case cluster.SubClusterForSuanli:
			// 后续实现
		case cluster.SubClusterForHunbu:
			hbSubClusterIdArr = append(hbSubClusterIdArr, subClusterId)
		default:
			nmSubClusterIdArr = append(nmSubClusterIdArr, subClusterId)
		}
	}

	if len(hbSubClusterIdArr) > 0 {
		hbSubClusterId = strings.Join(hbSubClusterIdArr, ",")
	}

	if len(nmSubClusterIdArr) > 0 {
		nmSubClusterId = strings.Join(nmSubClusterIdArr, ",")
	}

	return hbSubClusterId, nmSubClusterId, nil
}

// createFederationNamespace create ns
func createFederationNamespace(hostClusterId string, manager *FederationManager,
	req *federationmgr.CreateFederationClusterNamespaceRequest) error {

	selector := ""
	blog.Infof("createFederationNamespace hostClusterId: %s, req: %+v", hostClusterId, req)

	// 联邦子集群亲和性
	if req.ClusterAffinity != nil {
		req.Annotations[cluster.ClusterAffinityMode] = req.ClusterAffinity.Mode
		if req.ClusterAffinity.LabelSelector != nil {
			// 有matchLabels，用matchLabels; 没有matchLabels，用matchExpressions
			if len(req.ClusterAffinity.LabelSelector.MatchExpressions) > 0 {
				bytes, err := json.Marshal(req.ClusterAffinity.LabelSelector.MatchExpressions)
				if err != nil {
					blog.Errorf(
						"createFederationNamespace json.Marshal failed, matchExpressions: %+v, err: %s",
						req.ClusterAffinity.LabelSelector.MatchExpressions, err.Error())
					return err
				}
				req.Annotations[cluster.ClusterAffinitySelector] = string(bytes)
			} else if len(req.ClusterAffinity.LabelSelector.MatchLabels) > 0 {
				bytes, err := json.Marshal(req.ClusterAffinity.LabelSelector.MatchLabels)
				if err != nil {
					blog.Errorf(
						"createFederationNamespace json.Marshal failed, matchLabels: %+v, err: %s",
						req.ClusterAffinity.LabelSelector.MatchLabels, err.Error())
					return err
				}
				req.Annotations[cluster.ClusterAffinitySelector] = string(bytes)
			}

			selector = req.Annotations[cluster.ClusterAffinitySelector]
		}
	}

	req.Annotations[cluster.FedNamespaceIsFederatedKey] = cluster.ValueIsTrue
	// 将ns状态写入到annotations中
	req.Annotations[cluster.HostClusterNamespaceStatus] = cluster.NamespaceCreating
	// 1. 创建fedClusterNs
	err := manager.clusterCli.CreateClusterNamespace(hostClusterId, req.Namespace, req.Annotations)
	if err != nil {
		blog.Errorf(
			"CreateClusterNamespace failed,hostClusterId: %s,req: %+v, err: %s",
			hostClusterId, req, err.Error())
		return err
	}

	if selector != "" {
		delete(req.Annotations, cluster.ClusterAffinitySelector)
	}

	delete(req.Annotations, cluster.FedNamespaceIsFederatedKey)
	delete(req.Annotations, cluster.HostClusterNamespaceStatus)
	blog.Infof("createFederationNamespace successful")

	return nil
}

// DeleteFederationClusterNamespace delete ns
func (f *FederationManager) DeleteFederationClusterNamespace(ctx context.Context,
	req *federationmgr.DeleteFederationClusterNamespaceRequest,
	resp *federationmgr.DeleteFederationClusterNamespaceResponse) error {

	blog.Infof("Received BcsFederationManager.DeleteFederationClusterNamespace request, req: %+v", req)

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate DeleteFederationClusterNamespace request failed, err: %s",
			err.Error()))
	}

	// 根据联邦集群代理ID获取 联邦集群ID
	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), req.ClusterId)
	if err != nil {
		return ErrReturn(resp,
			fmt.Sprintf("DeleteFederationClusterNamespace get federation cluster from federationmanager failed, "+
				"clusterId: %s, err: %s", req.ClusterId, err.Error()))
	}
	// get namespace
	federationClusterNamespace, err := f.clusterCli.GetNamespace(fedCluster.HostClusterID, req.Namespace)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("GetNamespace failed, err: %s", err.Error()))
	}
	// delete namespace
	err = f.clusterCli.DeleteNamespace(fedCluster.HostClusterID, req.Namespace)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("DeleteNamespace failed, err: %s", err.Error()))
	}

	// 获取子集群范围
	subClusterIds := make([]string, 0)
	if clusterRangeStr, ok := federationClusterNamespace.Annotations[cluster.FedNamespaceClusterRangeKey]; ok {
		if len(clusterRangeStr) != 0 {
			lower := strings.Split(clusterRangeStr, ",")
			for _, sc := range lower {
				subClusterIds = append(subClusterIds, strings.ToUpper(sc))
			}
		}
	}

	// 获取混部集群范围
	reqMap := make(map[string]string)
	hbClusterIdStr, nmClusterIdStr, err := f.transferDeleteSubClusterNamespace(fedCluster.HostClusterID, subClusterIds)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("transferDeleteSubClusterNamespace failed, err: %s", err.Error()))
	}

	if hbClusterIdStr != "" {
		reqMap[cluster.SubClusterForHunbu] = hbClusterIdStr
	}

	if nmClusterIdStr != "" {
		reqMap[cluster.SubClusterForNormal] = nmClusterIdStr
	}

	if len(reqMap) == 0 {
		blog.Errorf("DeleteFederationClusterNamespace taskParams is empty, HostClusterID %s, subClusterIds %+v, "+
			"req %+v", fedCluster.HostClusterID, subClusterIds, req)
	} else {
		reqParameterBytes, err := json.Marshal(reqMap)
		if err != nil {
			return ErrReturn(resp, fmt.Sprintf("DeleteFederationClusterNamespace.Marshal failed,err: %s", err.Error()))
		}

		blog.Infof("DeleteFederationClusterNamespace task reqParameter: %s", string(reqParameterBytes))

		// 启动task
		t, err := fedtasks.NewHandleNamespaceQuotaTask(
			&fedtasks.HandleNamespaceQuotaOptions{
				HandleType:    cluster.DeleteKey,
				FedClusterId:  fedCluster.FederationClusterID,
				HostClusterId: fedCluster.HostClusterID,
				Namespace:     req.Namespace,
				Parameter:     string(reqParameterBytes),
			}).BuildTask(req.Operator)
		if err != nil {
			return ErrReturn(resp, fmt.Sprintf("BuildTask error when delete federation cluster namespace, err: %s", err.Error()))
		}
		if err = f.taskmanager.Dispatch(t); err != nil {
			return ErrReturn(resp, fmt.Sprintf("DeleteFederationClusterNamespace run task failed,err: %s", err.Error()))
		}
	}

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	return nil
}

// UpdateFederationClusterNamespace update ns
func (f *FederationManager) UpdateFederationClusterNamespace(ctx context.Context,
	req *federationmgr.UpdateFederationClusterNamespaceRequest,
	resp *federationmgr.UpdateFederationClusterNamespaceResponse) error {

	blog.Infof("Received BcsFederationManager.UpdateFederationClusterNamespace request, req: %+v", req)

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate UpdateFederationClusterNamespace request failed, err: %s", err.Error()))
	}

	// 根据联邦集群代理ID获取 联邦集群ID
	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), req.ClusterId)
	if err != nil {
		return ErrReturn(resp,
			fmt.Sprintf("UpdateFederationClusterNamespace get federation cluster from federationmanager failed, "+
				"clusterId: %s, err: %s", req.ClusterId, err.Error()))
	}

	subClusterIdMap := make(map[string]string)
	// 检测cluster-range中包含的子集群是否都是该联邦的有效子集群
	if clusterRangeStr, ok := req.Annotations[cluster.FedNamespaceClusterRangeKey]; ok {
		blog.Infof("UpdateFederationClusterNamespace has sub cluster: %+v", req.Annotations)
		if len(clusterRangeStr) != 0 {
			lower := strings.Split(clusterRangeStr, ",")
			for _, sc := range lower {
				subClusterId := strings.ToUpper(sc)
				// 检查子集群中是否存在命名空间
				// NOCC:vetshadow/shadow(设计如此:这里err可以被覆盖)
				subCluster, err := f.checkUpdateSubClusterNamespace(subClusterId, fedCluster.FederationClusterID)
				if err != nil {
					return ErrReturn(resp, fmt.Sprintf("checkCreateSubClusterNamespace failed, err: %s", err.Error()))
				}

				if subCluster != nil {
					subClusterIdMap[subClusterId] = subCluster.SubClusterName
				}
			}
		}
	}
	// update ns
	err = f.updateFederationNamespace(fedCluster, req, subClusterIdMap)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("updateFederationNamespace failed, err: %s", err.Error()))
	}

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	return nil
}

// updateFederationNamespace update ns
func (f *FederationManager) updateFederationNamespace(fedCluster *store.FederationCluster,
	req *federationmgr.UpdateFederationClusterNamespaceRequest, subClusterIdMap map[string]string) error {

	// 获取fedNamespace
	ns, err := f.clusterCli.GetNamespace(fedCluster.HostClusterID, req.Namespace)
	if err != nil {
		blog.Errorf("UpdateFederationClusterNamespace.GetNamespace failed clusterId: %s, namespace: %s, err: %s",
			fedCluster.HostClusterID, req.Namespace, err.Error())
		return fmt.Errorf("UpdateFederationClusterNamespace.GetNamespace failed, err: %s", err.Error())
	}

	// 更新annotations
	for key, val := range req.Annotations {
		ns.Annotations[key] = val
	}
	ns.Annotations[cluster.ClusterAffinityMode] = req.ClusterAffinity.Mode
	if req.ClusterAffinity.LabelSelector != nil {
		if len(req.ClusterAffinity.LabelSelector.MatchLabels) > 0 {
			marshal, err := json.Marshal(req.ClusterAffinity.LabelSelector.MatchLabels)
			if err != nil {
				return fmt.Errorf("UpdateFederationClusterNamespace.Marshal failed, err: %s", err.Error())
			}
			ns.Annotations[cluster.ClusterAffinitySelector] = string(marshal)
		} else if len(req.ClusterAffinity.LabelSelector.MatchExpressions) > 0 {
			marshal, err := json.Marshal(req.ClusterAffinity.LabelSelector.MatchExpressions)
			if err != nil {
				return fmt.Errorf("UpdateFederationClusterNamespace.Marshal failed, err: %s", err.Error())
			}
			ns.Annotations[cluster.ClusterAffinitySelector] = string(marshal)
		}
	}

	// 获取异步任务需要的参数
	reqMap, createNamespaceParamMap, err := f.transferUpdateSubClusterNamespaceParam(fedCluster.HostClusterID,
		subClusterIdMap, ns.Annotations, req)
	if err != nil {
		blog.Errorf("transfer sub cluster param failed err: %s", err.Error())
		return err
	}

	if len(reqMap) == 0 {
		blog.Errorf("UpdateFederationClusterNamespace taskParams is empty, HostClusterID %s, subClusterIds %+v, "+
			"req %+v", fedCluster.HostClusterID, subClusterIdMap, req)
		ns.Annotations[cluster.HostClusterNamespaceStatus] = cluster.NamespaceSuccess
		// update ns
		err = f.clusterCli.UpdateNamespace(fedCluster.HostClusterID, ns)
		if err != nil {
			blog.Errorf(
				"UpdateFederationClusterNamespace  failed clusterId: %s, namespace: %+v, err: %s",
				fedCluster.HostClusterID, ns, err.Error())
			return fmt.Errorf("UpdateFederationClusterNamespace.UpdateNamespace failed, err: %s", err.Error())
		}
	} else {
		reqParameterBytes, err := json.Marshal(reqMap)
		if err != nil {
			blog.Errorf("UpdateFederationClusterNamespace.Marshal failed,err: %s", err.Error())
			return err
		}

		// 启动task
		t, err := fedtasks.NewHandleNamespaceQuotaTask(&fedtasks.HandleNamespaceQuotaOptions{
			HandleType: cluster.UpdateKey, FedClusterId: fedCluster.FederationClusterID,
			HostClusterId: fedCluster.HostClusterID, Namespace: req.Namespace,
			Parameter: string(reqParameterBytes)}).BuildTask(req.Operator)
		if err != nil {
			blog.Errorf("BuildTask error when update federation cluster namespace, err: %s", err.Error())
			return err
		}
		ns.Annotations[cluster.HostClusterNamespaceStatus] = cluster.NamespaceCreating
		ns.Annotations[cluster.FederationClusterTaskIDLabelKey] = t.GetTaskID()
		// update ns
		err = f.clusterCli.UpdateNamespace(fedCluster.HostClusterID, ns)
		if err != nil {
			blog.Errorf(
				"UpdateFederationClusterNamespace  failed clusterId: %s, namespace: %+v, err: %s",
				fedCluster.HostClusterID, ns, err.Error())
			return fmt.Errorf("UpdateFederationClusterNamespace.UpdateNamespace failed, err: %s", err.Error())
		}

		if err = f.taskmanager.Dispatch(t); err != nil {
			blog.Errorf("UpdateFederationClusterNamespace run task failed,err: %s", err.Error())
			return err
		}
	}
	// build task for create ns
	if len(createNamespaceParamMap) > 0 {
		reqParameterBytes, err := json.Marshal(createNamespaceParamMap)
		if err != nil {
			blog.Errorf("UpdateFederationClusterNamespace.Marshal failed,err: %s", err.Error())
			return err
		}
		// 启动task
		t, err := fedtasks.NewHandleNamespaceQuotaTask(
			&fedtasks.HandleNamespaceQuotaOptions{
				HandleType: cluster.CreateKey, FedClusterId: fedCluster.FederationClusterID,
				HostClusterId: fedCluster.HostClusterID, Namespace: req.Namespace,
				Parameter: string(reqParameterBytes)}).BuildTask(req.Operator)
		if err != nil {
			blog.Errorf("BuildTask error when create federation cluster namespace, err: %s", err.Error())
			return err
		}
		if err = f.taskmanager.Dispatch(t); err != nil {
			blog.Errorf("UpdateFederationClusterNamespace run task failed,err: %s", err.Error())
			return err
		}
	}
	return nil
}

// getParams init params
func getParams(req *federationmgr.UpdateFederationClusterNamespaceRequest, annotations map[string]string) (
	bool, string, string) {

	// init params
	var location string
	var scheduleAlgorithm string
	var isPrivateResource string
	if req.Annotations[cluster.AnnotationIsPrivateResourceKey] == "" {
		isPrivateResource = annotations[cluster.AnnotationIsPrivateResourceKey]
	} else {
		isPrivateResource = req.Annotations[cluster.AnnotationIsPrivateResourceKey]
	}

	// scheduleAlgorithm
	if req.Annotations[cluster.AnnotationScheduleAlgorithmKey] == "" {
		scheduleAlgorithm = annotations[cluster.AnnotationScheduleAlgorithmKey]
	} else {
		scheduleAlgorithm = req.Annotations[cluster.AnnotationScheduleAlgorithmKey]
	}

	// bkbcs.tencent.com/taiji-location
	if req.Annotations[cluster.AnnotationSubClusterForTaiji] == "" {
		location = annotations[cluster.AnnotationSubClusterForTaiji]
	} else {
		location = req.Annotations[cluster.AnnotationSubClusterForTaiji]
	}

	var isPr bool
	if isPrivateResource == "true" {
		isPr = true
	}

	return isPr, scheduleAlgorithm, location
}

// getBkModuleParams init params
func getBkModuleParams(req *federationmgr.UpdateFederationClusterNamespaceRequest, annotations map[string]string) (
	string, string) {

	// init params
	var bkBizId string
	var bkModuleId string
	// bkbcs.tencent.com/bk-biz-id
	if req.Annotations[cluster.FedNamespaceBkBizId] == "" {
		bkBizId = annotations[cluster.FedNamespaceBkBizId]
	} else {
		bkBizId = req.Annotations[cluster.FedNamespaceBkBizId]
	}

	// bkbcs.tencent.com/bk-module-id
	if req.Annotations[cluster.FedNamespaceBkModuleId] == "" {
		bkModuleId = annotations[cluster.FedNamespaceBkModuleId]
	} else {
		bkModuleId = req.Annotations[cluster.FedNamespaceBkModuleId]
	}

	return bkBizId, bkModuleId
}
