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

// Package handler xxx
package handler

import (
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	federationmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/proto/bcs-federation-manager"
)

// ErrReturn return error and set code and message
func ErrReturn(resp interface{}, errStr string) error {
	v := reflect.ValueOf(resp)
	if v.Elem().FieldByName("Code").IsValid() {
		codePtr := IntToUint32Ptr(common.AdditionErrorCode + 500)
		v.Elem().FieldByName("Code").Set(reflect.ValueOf(codePtr))
	}

	if v.Elem().FieldByName("Message").IsValid() {
		v.Elem().FieldByName("Message").SetString(errStr)
	}
	// log
	blog.Error(errStr)
	return nil
}

// IsFederationNamespace check if namespace is federation namespace
func IsFederationNamespace(namespace *corev1.Namespace) bool {
	if namespace == nil {
		return false
	}

	// check is federated namespace
	isFedNamespace, ok := namespace.Annotations[cluster.FedNamespaceIsFederatedKey]
	if !ok || isFedNamespace != "true" {
		return false
	}

	return true
}

// TransferFedCluster transfer store.FederationCluster to federationmgr.FederationCluster
func TransferFedCluster(storeCluster *store.FederationCluster,
	storeSubClusters []*store.SubCluster, fedNamespaces []*cluster.FederationNamespace) *federationmgr.FederationCluster {

	// subcluster list
	fmSubClusters := make([]*federationmgr.FederationSubCluster, 0, len(storeSubClusters))
	for _, c := range storeSubClusters {
		fmSubClusters = append(fmSubClusters, transferToHandlerSubcluster(c, nil))
	}

	// federation namespaces list
	fmNamespaces := make([]*federationmgr.FederationNamespace, 0, len(fedNamespaces))
	for _, ns := range fedNamespaces {
		fmNamespaces = append(fmNamespaces, transferToHandlerFedNamespace(ns, storeCluster))
	}

	return &federationmgr.FederationCluster{
		FederationClusterId:   storeCluster.FederationClusterID,
		FederationClusterName: storeCluster.FederationClusterName,
		HostClusterId:         storeCluster.HostClusterID,
		ProjectCode:           storeCluster.ProjectCode,
		ProjectId:             storeCluster.ProjectID,
		CreatedTime:           storeCluster.CreatedTime.Format(time.RFC3339),
		UpdatedTime:           storeCluster.UpdatedTime.Format(time.RFC3339),
		Status:                storeCluster.Status,
		SubClusters:           fmSubClusters,
		FederationNamespaces:  fmNamespaces,
	}
}
