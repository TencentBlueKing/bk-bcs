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

package portpoolcontroller

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	ingresscommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/portpoolcache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/utils"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
)

// PortPoolHandler handler for port pool
type PortPoolHandler struct {
	namespace   string
	region      string
	k8sClient   client.Client
	lbClient    cloud.LoadBalance
	poolCache   *portpoolcache.Cache
	lbIDCache   *gocache.Cache
	lbNameCache *gocache.Cache
}

// newPortPoolHandler create port pool handler
func newPortPoolHandler(ns, region string,
	lbClient cloud.LoadBalance, k8sClient client.Client, poolCache *portpoolcache.Cache,
	lbIDCache *gocache.Cache, lbNameCache *gocache.Cache) *PortPoolHandler {
	return &PortPoolHandler{
		namespace:   ns,
		region:      region,
		k8sClient:   k8sClient,
		lbClient:    lbClient,
		poolCache:   poolCache,
		lbIDCache:   lbIDCache,
		lbNameCache: lbNameCache,
	}
}

// ensure port pool
// the returned bool value indicates whether you need to retry
// nolint  funlen
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
		ListenerAttr:  pool.Spec.ListenerAttribute,
		lbIDCache:     pph.lbIDCache,
		lbNameCache:   pph.lbNameCache}

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
	pool.Status.PoolItemStatuses = append(pool.Status.PoolItemStatuses, newItemStatusList...)

	pool.Status.Status = checkPortPoolStatus(pool)

	err := pph.k8sClient.Status().Update(context.Background(), pool, &client.UpdateOptions{})
	if err != nil {
		return true, fmt.Errorf("update %s/%s status failed, err %s", pool.GetNamespace(), pool.GetName(), err.Error())
	}

	// if portItem.external changed, update related portBinding
	if err := pph.ensurePortBinding(pool); err != nil {
		return true, errors.Wrapf(err, "pool[%s/%s] ensurePortBinding failed", pool.GetNamespace(), pool.GetName())
	}

	// update related cache
	poolKey := ingresscommon.GetNamespacedNameKey(pool.GetName(), pool.GetNamespace())
	pph.ensureCache(poolKey, pool.GetAllocatePolicy(), tmpItemsStatus, successDeletedKeyMap, failedDeletedKeyMap,
		newItemStatusList, updateItemStatusMap)

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
		ListenerAttr:  pool.Spec.ListenerAttribute,
		lbIDCache:     pph.lbIDCache,
		lbNameCache:   pph.lbNameCache}

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
				pph.poolCache.SetPortPoolItemStatus(poolKey, itemStatus) // nolint
			}
		} else {
			pph.poolCache.DeletePortPoolItem(poolKey, itemKey)
		}
	}

	return true, nil
}

// if poolItem.external changed, update related portBinding
func (pph *PortPoolHandler) ensurePortBinding(pool *networkextensionv1.PortPool) error {
	for i, poolItem := range pool.Spec.PoolItems {
		portBindingList := &networkextensionv1.PortBindingList{}
		labelKey := utils.GenPortBindingLabel(pool.GetName(), pool.GetNamespace())
		if err := pph.k8sClient.List(context.Background(), portBindingList,
			client.MatchingLabels{labelKey: poolItem.ItemName}); err != nil {
			return errors.Wrapf(err, "list portBinding with label['%s'='%s'] failed", labelKey, poolItem.ItemName)
		}

		poolStatus := pool.Status.PoolItemStatuses[i]
		for _, portBinding := range portBindingList.Items {
			changed := false
			cpPortBinding := portBinding.DeepCopy()
			for idx, portBindingItem := range cpPortBinding.Spec.PortBindingList {
				// check if same item
				if portBindingItem.GetKey() != poolItem.GetKey() ||
					portBindingItem.PoolName != pool.Name ||
					portBindingItem.PoolNamespace != pool.Namespace {
					continue
				}
				// check if external changed
				if portBindingItem.External != poolItem.External {
					cpPortBinding.Spec.PortBindingList[idx].External = poolItem.External
					changed = true
				}
				// 支持用户新增LBID
				if !reflect.DeepEqual(portBindingItem.LoadBalancerIDs, poolItem.LoadBalancerIDs) {
					cpPortBinding.Spec.PortBindingList[idx].LoadBalancerIDs = poolItem.LoadBalancerIDs
					changed = true
				}
				if !reflect.DeepEqual(portBindingItem.PoolItemLoadBalancers, poolStatus.PoolItemLoadBalancers) {
					cpPortBinding.Spec.PortBindingList[idx].PoolItemLoadBalancers = poolStatus.PoolItemLoadBalancers
					changed = true
				}
			}

			if changed {
				blog.Infof("pool[%s/%s].poolItem[%s] changed, update related portBinding", pool.GetNamespace(),
					pool.GetName(), poolItem.ItemName)
				if err := pph.k8sClient.Update(context.Background(), cpPortBinding,
					&client.UpdateOptions{}); err != nil {
					return errors.Wrapf(err, "update portBinding[%s/%s] failed", cpPortBinding.GetNamespace(),
						cpPortBinding.GetName())
				}
			}
		}
	}

	return nil
}

func (pph *PortPoolHandler) ensureCache(poolKey string, allocatePolicy string, tmpItemsStatus []*networkextensionv1.
	PortPoolItemStatus,
	successDeletedKeyMap, failedDeletedKeyMap map[string]struct{}, newItemStatusList []*networkextensionv1.
		PortPoolItemStatus, updateItemStatusMap map[string]*networkextensionv1.PortPoolItemStatus) {

	pph.poolCache.Lock()
	defer pph.poolCache.Unlock()

	for _, itemStatus := range tmpItemsStatus {
		itemKey := itemStatus.GetKey()
		if _, ok := successDeletedKeyMap[itemKey]; !ok {
			if _, inOk := failedDeletedKeyMap[itemKey]; inOk {
				pph.poolCache.SetPortPoolItemStatus(poolKey, itemStatus) // nolint
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
		if err := pph.poolCache.AddPortPoolItem(poolKey, allocatePolicy, itemStatus); err != nil {
			blog.Warnf("failed to add port pool %s item %v to cache, err %s", poolKey, itemStatus, err.Error())
		} else {
			blog.Infof("add port pool %s item %v to cache", poolKey, itemStatus)
		}
	}
	// update item status
	for _, itemStatus := range updateItemStatusMap {
		pph.poolCache.SetPortPoolItemStatus(poolKey, itemStatus) // nolint
		blog.Infof("set port pool %s item %s status to %s", poolKey, itemStatus.ItemName, itemStatus.Status)
	}
}
