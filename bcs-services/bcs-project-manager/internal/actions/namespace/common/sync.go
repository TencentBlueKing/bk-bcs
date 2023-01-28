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

package common

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/config"
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
	defer session.Close()
	prefix := fmt.Sprintf("%s/%s/%s", config.NamespaceSyncLockPrefix, projectCode, clusterID)
	mu := concurrency.NewMutex(session, prefix)
	timeoutCtx, _ := context.WithTimeout(context.TODO(), time.Second)
	err = mu.Lock(timeoutCtx)
	if err != nil {
		logging.Error("tryLock prefix %s with unexpected err: %s", prefix, err.Error())
		return nil
	}
	defer mu.Unlock(context.TODO())
	cluster, err := clustermanager.GetCluster(clusterID)
	if err != nil {
		logging.Error("get cluster %s from cluster-manager failed, err: %s", clusterID, err.Error())
		return err
	}
	creator := cluster.GetCreator()
	if err != nil {
		return errorx.NewClusterErr(err.Error())
	}
	ccNsList, err := bcscc.ListNamespaces(projectCode, clusterID)
	if err != nil {
		return errorx.NewRequestBCSCCErr(err.Error())
	}
	// insert new namespace to bcscc
	ccnsMap := map[string]bcscc.NamespaceData{}
	for _, ccns := range ccNsList.Results {
		ccnsMap[ccns.Name] = ccns
	}
	g1, ctx := errgroup.WithContext(context.Background())
	for _, item := range namespaces {
		if _, ok := ccnsMap[item.GetName()]; !ok {
			ns := item
			g1.Go(func() error {
				if err := bcscc.CreateNamespace(projectCode, clusterID, ns.GetName(), creator); err != nil {
					return errorx.NewRequestBCSCCErr(err.Error())
				}
				return nil
			})
		}
	}
	// delete old namespace in bcscc
	bcsnsMap := map[string]corev1.Namespace{}
	for _, item := range namespaces {
		bcsnsMap[item.GetName()] = item
	}
	g2, ctx := errgroup.WithContext(ctx)
	for _, item := range ccNsList.Results {
		if _, ok := bcsnsMap[item.Name]; !ok {
			ns := item
			g2.Go(func() error {
				if err := bcscc.DeleteNamespace(projectCode, clusterID, ns.Name); err != nil {
					return errorx.NewRequestBCSCCErr(err.Error())
				}
				return nil
			})
		}
	}
	if err := g1.Wait(); err != nil {
		logging.Error("create namespace in bcscc failed, err:%s", err.Error())
		return err
	}
	if err := g2.Wait(); err != nil {
		logging.Error("delete namespace in bcscc failed, err:%s", err.Error())
		return err
	}
	logging.Info("sync namespace in %s/%s success", projectCode, clusterID)
	return nil
}
