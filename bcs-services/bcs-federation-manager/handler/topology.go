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

// Package handler topology service
package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	federationmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/proto/bcs-federation-manager"
)

// ListFederationClusterWithSubcluster list federation clusters which contains sub clusters
func (f *FederationManager) ListFederationClusterWithSubcluster(ctx context.Context,
	req *federationmgr.ListFederationClusterWithSubclusterRequest,
	resp *federationmgr.ListFederationClusterWithSubclusterResponse) error {

	blog.Infof("Received BcsFederationManager.ListFederationClusterWithSubcluster request, "+
		"conditions: %v, sub_conditions: %v",
		req.GetConditions(), req.GetSubConditions())

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate ListFederationClusterWithSubcluster request failed, err: %s",
			err.Error()))
	}

	// get federation clusters
	fedClusters, err := f.store.ListFederationClusters(ctx, &store.FederationListOptions{
		Conditions: req.GetConditions(),
	})
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("ListFederationClusterWithSubcluster error when get fed clsuters: %s",
			err.Error()))
	}

	// match subcluster by federation cluster
	data := []*federationmgr.FederationClusterWithSubcluster{}
	for _, fc := range fedClusters {
		fedInfo, err := f.getFedWithSubcluster(ctx, fc, req.GetSubConditions())
		if err != nil {
			return ErrReturn(resp, fmt.Sprintf("ListFederationClusterWithSubcluster error when get sub clusters: %s, "+
				"error: %s", fc.HostClusterID, err.Error()))
		}
		data = append(data, fedInfo)
	}

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = data
	return nil
}

// ListFederationClusterWithNamespace list federation clusters which contains federation namespaces
func (f *FederationManager) ListFederationClusterWithNamespace(ctx context.Context,
	req *federationmgr.ListFederationClusterWithNamespaceRequest,
	resp *federationmgr.ListFederationClusterWithNamespaceResponse) error {

	blog.Infof("Received BcsFederationManager.ListFederationClusterWithNamespace request, "+
		"conditions: %v, sub_conditions: %v",
		req.GetConditions(), req.GetSubConditions())

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate ListFederationClusterWithSubcluster request failed, err: %s",
			err.Error()))
	}

	// get federation clusters
	fedClusters, err := f.store.ListFederationClusters(ctx, &store.FederationListOptions{
		Conditions: req.GetConditions(),
	})
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("ListFederationClusterWithNamespace error when get fed clsuters: %s",
			err.Error()))
	}

	// fill federation namespaces by federation cluster
	data := []*federationmgr.FederationClusterWithNamespace{}
	for _, fc := range fedClusters {
		// get federation namespaces
		// todo 待添加命名空间的过滤，根据project_code进行过滤
		namespaceList, err := f.clusterCli.ListFederationNamespaces(fc.HostClusterID)
		if err != nil {
			return ErrReturn(resp, fmt.Sprintf("ListFederationClusterWithNamespace error "+
				"when get federation namespaces: %s", err.Error()))
		}
		if len(namespaceList) == 0 {
			blog.Infof("namespaceList is empty for federation cluster [%s]", fc.FederationClusterID)
		}

		federationNamespaces := make([]*federationmgr.FederationNamespace, 0)
		for _, ns := range namespaceList {
			federationNamespaces = append(federationNamespaces, transferToHandlerFedNamespace(ns, fc))
		}
		data = append(data, transferToHandlerFedClusterWithNamespaces(fc, federationNamespaces))
	}

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = data
	return nil
}

// NOCC:tosa/fn_length(设计如此)
func transferToHandlerFedClusterWithSubclusters(original *store.FederationCluster,
	subclusters []*federationmgr.FederationSubCluster) *federationmgr.FederationClusterWithSubcluster {
	return &federationmgr.FederationClusterWithSubcluster{
		ProjectCode:           original.ProjectCode,
		ProjectId:             original.ProjectID,
		FederationClusterId:   original.FederationClusterID,
		FederationClusterName: original.FederationClusterName,
		HostClusterId:         original.HostClusterID,
		CreatedTime:           original.CreatedTime.Format(time.RFC3339),
		UpdatedTime:           original.UpdatedTime.Format(time.RFC3339),
		SubClusters:           subclusters,
	}
}

func transferToHandlerSubcluster(original *store.SubCluster, namespaces []string) *federationmgr.FederationSubCluster {
	return &federationmgr.FederationSubCluster{
		ProjectCode:                original.ProjectCode,
		ProjectId:                  original.ProjectID,
		SubClusterId:               original.SubClusterID,
		SubClusterName:             original.SubClusterName,
		FederationClusterId:        original.FederationClusterID,
		HostClusterId:              original.HostClusterID,
		ClusternetClusterName:      original.ClusternetClusterName,
		ClusternetClusterNamespace: original.ClusternetClusterNamespace,
		CreatedTime:                original.CreatedTime.Format(time.RFC3339),
		UpdatedTime:                original.UpdatedTime.Format(time.RFC3339),
		Status:                     original.Status,
		Labels:                     original.Labels,
		FederationNamespaces:       namespaces,
	}
}

// mapSubclusterToNamespaces return a map which map [subclusterID] to namespace list
func mapSubclusterToNamespaces(subclusters []*store.SubCluster,
	federationNamespaces []*cluster.FederationNamespace) map[string][]string {
	idNssMaps := make(map[string][]string, 0)
	if len(subclusters) == 0 || len(federationNamespaces) == 0 {
		return idNssMaps
	}

	for _, ns := range federationNamespaces {
		// Note: If the list of sub clusters in a namespace is empty,
		// then the effective range of the namespace is all sub clusters
		if len(ns.SubClusters) == 0 {
			for _, cluster := range subclusters {
				ns.SubClusters = append(ns.SubClusters, cluster.SubClusterID)
			}
		}
		// If there is a specified sub cluster, add the specified subset group
		for _, id := range ns.SubClusters {
			idNssMaps[id] = append(idNssMaps[id], ns.Namespace)
		}
	}
	return idNssMaps
}

// NOCC:tosa/fn_length(设计如此)
func transferToHandlerFedClusterWithNamespaces(original *store.FederationCluster,
	federationNamespaces []*federationmgr.FederationNamespace) *federationmgr.FederationClusterWithNamespace {
	return &federationmgr.FederationClusterWithNamespace{
		ProjectCode:           original.ProjectCode,
		ProjectId:             original.ProjectID,
		FederationClusterId:   original.FederationClusterID,
		FederationClusterName: original.FederationClusterName,
		HostClusterId:         original.HostClusterID,
		CreatedTime:           original.CreatedTime.Format(time.RFC3339),
		UpdatedTime:           original.UpdatedTime.Format(time.RFC3339),
		FederationNamespaces:  federationNamespaces,
	}
}

func transferToHandlerFedNamespace(original *cluster.FederationNamespace,
	federationCluster *store.FederationCluster) *federationmgr.FederationNamespace {

	//todo: 这里直接用的联邦命名空间的label的子集群列表，
	// 如果这个联邦命名空间的label是空的话，这里会是空
	// 正确的情况应该是如果是空的列表，则应该作用于全部子集群

	return &federationmgr.FederationNamespace{
		ProjectCode:         original.ProjectCode,
		FederationNamespace: original.Namespace,
		CreatedTime:         original.CreatedTime.Format(time.RFC3339),
		SubClusters:         original.SubClusters,
		FederationClusterId: federationCluster.FederationClusterID,
		HostClusterId:       federationCluster.HostClusterID,
	}
}

func (f *FederationManager) getFedWithSubcluster(ctx context.Context, fc *store.FederationCluster,
	conditions map[string]string) (*federationmgr.FederationClusterWithSubcluster, error) {

	// get sub clusters
	storeSubClusters, err := f.store.ListSubClusters(ctx, &store.SubClusterListOptions{
		FederationClusterID: fc.FederationClusterID,
		Conditions:          conditions,
	})
	if err != nil {
		return nil, err
	}

	if len(storeSubClusters) == 0 {
		blog.Infof("subcluster is empty for federation cluster [%s], sub_conditions: %v",
			fc.FederationClusterID, conditions)
		return transferToHandlerFedClusterWithSubclusters(fc, make([]*federationmgr.FederationSubCluster, 0)), nil
	}

	// get federation namespaces
	namespaceList, err := f.clusterCli.ListFederationNamespaces(fc.HostClusterID)
	if err != nil {
		return nil, fmt.Errorf("get federation namespaces error: %s", err.Error())
	}

	// Map the sub-clusters to a list of namespaces, which can be scheduled to the corresponding sub-clusters.
	idNssMaps := mapSubclusterToNamespaces(storeSubClusters, namespaceList)
	subClusters := make([]*federationmgr.FederationSubCluster, 0)
	for _, sc := range storeSubClusters {
		subClusters = append(subClusters, transferToHandlerSubcluster(sc, idNssMaps[sc.SubClusterID]))
	}
	return transferToHandlerFedClusterWithSubclusters(fc, subClusters), nil
}
