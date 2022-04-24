/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package portpoolcontroller

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	ingresscommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/portpoolcache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PortPoolHandler handler for port pool
type PortPoolHandler struct {
	namespace string
	region    string
	k8sClient client.Client
	lbClient  cloud.LoadBalance
	poolCache *portpoolcache.Cache
}

// newPortPoolHandler create port pool handler
func newPortPoolHandler(ns, region string,
	lbClient cloud.LoadBalance, k8sClient client.Client, poolCache *portpoolcache.Cache) *PortPoolHandler {
	return &PortPoolHandler{
		namespace: ns,
		region:    region,
		k8sClient: k8sClient,
		lbClient:  lbClient,
		poolCache: poolCache,
	}
}

// ensure port pool
// the returned bool value indicates whether you need to retry
func (pph *PortPoolHandler) ensurePortPool(pool *networkextensionv1.PortPool) (bool, error) {
	defItemMap := make(map[string]*networkextensionv1.PortPoolItem)
	for _, poolItem := range pool.Spec.PoolItems {
		defItemMap[poolItem.GetKey()] = poolItem
	}
	activeItemMap := make(map[string]*networkextensionv1.PortPoolItemStatus)
	for _, poolItemStatus := range pool.Status.PoolItemStatuses {
		activeItemMap[poolItemStatus.GetKey()] = poolItemStatus
	}

	// item to delete
	var delItemsStatus []*networkextensionv1.PortPoolItemStatus
	for k := range activeItemMap {
		if _, ok := defItemMap[k]; !ok {
			delItemsStatus = append(delItemsStatus, activeItemMap[k])
		}
	}

	poolItemHandler := &PortPoolItemHandler{
		PortPoolName:  pool.Name,
		Namespace:     pph.namespace,
		DefaultRegion: pph.region,
		LbClient:      pph.lbClient,
		K8sClient:     pph.k8sClient,
		ListenerAttr:  pool.Spec.ListenerAttribute}

	pph.poolCache.Lock()
	defer pph.poolCache.Unlock()

	// try to delete
	successDeletedKeyMap := make(map[string]struct{})
	failedDeletedKeyMap := make(map[string]struct{})
	for _, delItemStatus := range delItemsStatus {
		err := poolItemHandler.checkPortPoolItemDeletion(delItemStatus)
		if err != nil {
			blog.Warnf("cannot delete active item %s, err %s", delItemStatus.ItemName, err.Error())
			failedDeletedKeyMap[delItemStatus.GetKey()] = struct{}{}
		} else {
			successDeletedKeyMap[delItemStatus.GetKey()] = struct{}{}
		}
	}
	// delete from port pool status
	tmpItemsStatus := pool.Status.PoolItemStatuses
	pool.Status.PoolItemStatuses = make([]*networkextensionv1.PortPoolItemStatus, 0)
	for _, itemStatus := range tmpItemsStatus {
		if _, ok := successDeletedKeyMap[itemStatus.GetKey()]; !ok {
			if _, inOk := failedDeletedKeyMap[itemStatus.GetKey()]; inOk {
				itemStatus.Status = constant.PortPoolItemStatusDeleting
			}
			pool.Status.PoolItemStatuses = append(pool.Status.PoolItemStatuses, itemStatus)
		}
	}

	shouldRetry := false
	// try to add or update port pool item
	newItemStatusList := make([]*networkextensionv1.PortPoolItemStatus, 0)
	updateItemStatusMap := make(map[string]*networkextensionv1.PortPoolItemStatus)
	for _, tmpItem := range pool.Spec.PoolItems {
		var updateItemStatus *networkextensionv1.PortPoolItemStatus
		var retry bool
		tmpItemStatus, ok := activeItemMap[tmpItem.GetKey()]
		if !ok {
			updateItemStatus, retry = poolItemHandler.ensurePortPoolItem(tmpItem, nil)
			newItemStatusList = append(newItemStatusList, updateItemStatus)
		} else {
			updateItemStatus, retry = poolItemHandler.ensurePortPoolItem(tmpItem, tmpItemStatus)
			updateItemStatusMap[updateItemStatus.GetKey()] = updateItemStatus
		}
		if retry {
			shouldRetry = true
		}
	}
	for i, ts := range pool.Status.PoolItemStatuses {
		if _, ok := updateItemStatusMap[ts.GetKey()]; ok {
			pool.Status.PoolItemStatuses[i] = updateItemStatusMap[ts.GetKey()]
		}
	}
	for _, ts := range newItemStatusList {
		pool.Status.PoolItemStatuses = append(pool.Status.PoolItemStatuses, ts)
	}

	err := pph.k8sClient.Status().Update(context.Background(), pool, &client.UpdateOptions{})
	if err != nil {
		return true, fmt.Errorf("update %s/%s status failed, err %s", pool.GetNamespace(), pool.GetName(), err.Error())
	}

	// delete item from pool cache
	poolKey := ingresscommon.GetNamespacedNameKey(pool.GetName(), pool.GetNamespace())
	for _, itemStatus := range tmpItemsStatus {
		itemKey := itemStatus.GetKey()
		if _, ok := successDeletedKeyMap[itemKey]; !ok {
			if _, inOk := failedDeletedKeyMap[itemKey]; inOk {
				pph.poolCache.SetPortPoolItemStatus(poolKey, itemStatus)
				blog.Infof("set port pool %s item %s status to %s",
					poolKey, itemStatus.ItemName, constant.PortPoolItemStatusDeleting)
			}
		} else {
			pph.poolCache.DeletePortPoolItem(poolKey, itemKey)
			blog.Infof("delete port pool %s item %s", poolKey, itemStatus.ItemName)
		}
	}
	// add item to pool cache
	for _, itemStatus := range newItemStatusList {
		if err := pph.poolCache.AddPortPoolItem(poolKey, itemStatus); err != nil {
			blog.Warnf("failed to add port pool %s item %v to cache, err %s", poolKey, itemStatus, err.Error())
		} else {
			blog.Infof("add port pool %s item %v to cache", poolKey, itemStatus)
		}
	}
	// update item status
	for _, itemStatus := range updateItemStatusMap {
		pph.poolCache.SetPortPoolItemStatus(poolKey, itemStatus)
		blog.Infof("set port pool %s item %s status to %s", poolKey, itemStatus.ItemName, itemStatus.Status)
	}

	if len(failedDeletedKeyMap) != 0 || shouldRetry {
		return true, nil
	}
	return false, nil
}

// OnDelete delete port pool
func (pph *PortPoolHandler) deletePortPool(pool *networkextensionv1.PortPool) (bool, error) {
	if len(pool.Status.PoolItemStatuses) == 0 {
		pool.Finalizers = common.RemoveString(pool.Finalizers, constant.FinalizerNameBcsIngressController)
		if err := pph.k8sClient.Update(context.Background(), pool, &client.UpdateOptions{}); err != nil {
			return true, fmt.Errorf("removing finalizer from pool %s/%s failed, err %s",
				pool.GetNamespace(), pool.GetName(), err.Error())
		}
		return false, nil
	}

	poolItemHandler := &PortPoolItemHandler{
		PortPoolName:  pool.Name,
		Namespace:     pph.namespace,
		DefaultRegion: pph.region,
		LbClient:      pph.lbClient,
		K8sClient:     pph.k8sClient,
		ListenerAttr:  pool.Spec.ListenerAttribute}

	pph.poolCache.Lock()
	defer pph.poolCache.Unlock()

	// collect delete results
	successDeletedKeyMap := make(map[string]struct{})
	failedDeletedKeyMap := make(map[string]struct{})
	for _, itemStatus := range pool.Status.PoolItemStatuses {
		blog.V(3).Infof("check port pool item %s", itemStatus.ItemName)
		err := poolItemHandler.checkPortPoolItemDeletion(itemStatus)
		if err != nil {
			blog.Warnf("cannot delete active item %s, err %s", itemStatus.ItemName, err.Error())
			failedDeletedKeyMap[itemStatus.GetKey()] = struct{}{}
		} else {
			successDeletedKeyMap[itemStatus.GetKey()] = struct{}{}
		}
	}

	// delete from port pool status
	tmpItemsStatus := pool.Status.PoolItemStatuses
	pool.Status.PoolItemStatuses = make([]*networkextensionv1.PortPoolItemStatus, 0)
	for _, itemStatus := range tmpItemsStatus {
		if _, ok := successDeletedKeyMap[itemStatus.GetKey()]; !ok {
			if _, inOk := failedDeletedKeyMap[itemStatus.GetKey()]; inOk {
				itemStatus.Status = constant.PortPoolItemStatusDeleting
			}
			pool.Status.PoolItemStatuses = append(pool.Status.PoolItemStatuses, itemStatus)
		}
	}
	if err := pph.k8sClient.Status().Update(context.Background(), pool, &client.UpdateOptions{}); err != nil {
		return true, fmt.Errorf("update pool %s/%s failed when port pool was deleted, err %s",
			pool.GetNamespace(), pool.GetName(), err.Error())
	}

	// delete item from pool cache
	poolKey := ingresscommon.GetNamespacedNameKey(pool.GetName(), pool.GetNamespace())
	for _, itemStatus := range tmpItemsStatus {
		itemKey := itemStatus.GetKey()
		if _, ok := successDeletedKeyMap[itemKey]; !ok {
			if _, inOk := failedDeletedKeyMap[itemKey]; inOk {
				pph.poolCache.SetPortPoolItemStatus(poolKey, itemStatus)
			}
		} else {
			pph.poolCache.DeletePortPoolItem(poolKey, itemKey)
		}
	}

	return true, nil
}
