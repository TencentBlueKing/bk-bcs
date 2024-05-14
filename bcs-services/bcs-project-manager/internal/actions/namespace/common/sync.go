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

// Package common xxx
package common

import (
	"context"
	"fmt"
	"time"

	"go.etcd.io/etcd/client/v3/concurrency"
	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bcscc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/etcd"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
)

// SyncNamespace sync namespaces in paas-cc with apiserver
func SyncNamespace(projectCode, clusterID string, namespaces []corev1.Namespace) error {
	etcdCli, err := etcd.GetClient()
	if err != nil {
		logging.Error("get etcd client failed, err: %s", err.Error())
		return err
	}
	session, err := concurrency.NewSession(etcdCli, concurrency.WithTTL(3))
	if err != nil {
		logging.Error("new etcd session failed, err: %s", err.Error())
		return err
	}
	// nolint
	defer session.Close()
	prefix := fmt.Sprintf("%s/%s/%s", constant.NamespaceSyncLockPrefix, projectCode, clusterID)
	mu := concurrency.NewMutex(session, prefix)
	// NOCC:vet/vet(设计如此:)
	// nolint
	timeoutCtx, _ := context.WithTimeout(context.TODO(), time.Second)
	err = mu.Lock(timeoutCtx)
	if err != nil {
		logging.Error("tryLock prefix %s with unexpected err: %s", prefix, err.Error())
		return nil
	}
	// nolint
	defer mu.Unlock(context.TODO())
	cluster, err := clustermanager.GetCluster(clusterID)
	if err != nil {
		logging.Error("get cluster %s from cluster-manager failed, err: %s", clusterID, err.Error())
		return err
	}
	creator := cluster.GetCreator()
	ccNsList, err := bcscc.ListNamespaces(projectCode, clusterID)
	if err != nil {
		return errorx.NewRequestBCSCCErr(err.Error())
	}
	// insert new namespace to bcscc
	ccnsMap := map[string]bcscc.NamespaceData{}
	for _, ccns := range ccNsList.Results {
		ccnsMap[ccns.Name] = ccns
	}
	for _, item := range namespaces {
		if _, ok := ccnsMap[item.GetName()]; !ok {
			if err := bcscc.CreateNamespace(projectCode, clusterID, item.GetName(), creator); err != nil {
				return errorx.NewRequestBCSCCErr(err.Error())
			}
		}
	}
	// delete old namespace in bcscc
	bcsnsMap := map[string]corev1.Namespace{}
	for _, item := range namespaces {
		bcsnsMap[item.GetName()] = item
	}
	for _, item := range ccNsList.Results {
		if _, ok := bcsnsMap[item.Name]; !ok {
			if err := bcscc.DeleteNamespace(projectCode, clusterID, item.Name); err != nil {
				return errorx.NewRequestBCSCCErr(err.Error())
			}
		}
	}
	logging.Info("sync namespace in %s/%s success", projectCode, clusterID)
	return nil
}
