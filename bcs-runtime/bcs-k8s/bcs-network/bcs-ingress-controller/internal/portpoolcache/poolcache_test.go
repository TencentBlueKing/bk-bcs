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

package portpoolcache

import (
	"reflect"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func getNewCache() (*Cache, error) {
	cache := NewCache()
	if err := cache.AddPortPoolItem("test1.ns1", &networkextensionv1.PortPoolItemStatus{
		ItemName:        "item1",
		LoadBalancerIDs: []string{"lb1", "lb2"},
		StartPort:       30000,
		EndPort:         31000,
		Status:          constant.PortBindingStatusReady,
	}); err != nil {
		return nil, err
	}
	if err := cache.AddPortPoolItem("test1.ns1", &networkextensionv1.PortPoolItemStatus{
		ItemName:        "item2",
		LoadBalancerIDs: []string{"lb3", "lb4"},
		StartPort:       30000,
		EndPort:         31000,
		Status:          constant.PortBindingStatusReady,
	}); err != nil {
		return nil, err
	}
	return cache, nil
}

// TestCache tests cache functions
func TestCacheAllocate(t *testing.T) {
	cache, err := getNewCache()
	if err != nil {
		t.Fatalf("failed to get new cache")
	}

	_, cachePortItem, err := cache.AllocatePortBinding("test1.ns1", "TCP")
	if err != nil {
		t.Fatalf("allocate port binding failed, err %s", err.Error())
	}
	expectItem := AllocatedPortItem{
		PoolKey:     "test1.ns1",
		PoolItemKey: networkextensionv1.GetPoolItemKey("item1", []string{"lb1", "lb2"}),
		Protocol:    "TCP",
		StartPort:   30000,
		EndPort:     0,
		IsUsed:      true,
	}
	if !reflect.DeepEqual(cachePortItem, expectItem) {
		t.Fatalf("expect %v, but get %v", expectItem, cachePortItem)
	}

	_, cachePortItem, err = cache.AllocatePortBinding("test1.ns1", "TCP")
	if err != nil {
		t.Fatalf("allocate port binding failed, err %s", err.Error())
	}
	expectItem = AllocatedPortItem{
		PoolKey:     "test1.ns1",
		PoolItemKey: networkextensionv1.GetPoolItemKey("item1", []string{"lb1", "lb2"}),
		Protocol:    "TCP",
		StartPort:   30001,
		EndPort:     0,
		IsUsed:      true,
	}
	if !reflect.DeepEqual(cachePortItem, expectItem) {
		t.Fatalf("expect %v, but get %v", expectItem, cachePortItem)
	}

	cache.ReleasePortBinding("test1.ns1", networkextensionv1.GetPoolItemKey("item1", []string{"lb1", "lb2"}), "TCP", 30001, 0)

	if cache.portPoolMap["test1.ns1"].ItemList[0].PortListMap["TCP"].AllocatedPortNum != 1 {
		t.Fatalf("allocated port number %d is not 1",
			cache.portPoolMap["test1.ns1"].ItemList[0].PortListMap["TCP"].AllocatedPortNum)
	}
}

// TestDeletePortPoolItem test delete port pool item
func TestDeletePortPoolItem(t *testing.T) {
	cache, err := getNewCache()
	if err != nil {
		t.Fatalf("failed to get new cache")
	}

	testItem := &networkextensionv1.PortPoolItemStatus{
		ItemName:        "item2",
		LoadBalancerIDs: []string{"lb3", "lb4"},
		StartPort:       30000,
		EndPort:         31000,
		Status:          constant.PortBindingStatusReady,
	}
	cache.DeletePortPoolItem("test1.ns1", networkextensionv1.GetPoolItemKey("item1", []string{"lb1", "lb2"}))
	if !reflect.DeepEqual(cache.portPoolMap["test1.ns1"].ItemList[0].ItemStatus, testItem) {
		t.Fatalf("expect %v but get %v", testItem, cache.portPoolMap["test1.ns1"].ItemList[0].ItemStatus)
	}
}

// TestSetPortBindingUsed test set port binding used
func TestSetPortBindingUsed(t *testing.T) {
	cache, err := getNewCache()
	if err != nil {
		t.Fatalf("failed to get new cache")
	}

	cache.SetPortBindingUsed("test1.ns1", networkextensionv1.GetPoolItemKey("item1", []string{"lb1", "lb2"}), "TCP", 30000, 0)
	_, cachePortItem, err := cache.AllocatePortBinding("test1.ns1", "TCP")
	if err != nil {
		t.Fatalf("allocate port binding failed, err %s", err.Error())
	}
	expectItem := AllocatedPortItem{
		PoolKey:     "test1.ns1",
		PoolItemKey: networkextensionv1.GetPoolItemKey("item1", []string{"lb1", "lb2"}),
		Protocol:    "TCP",
		StartPort:   30001,
		EndPort:     0,
		IsUsed:      true,
	}
	if !reflect.DeepEqual(cachePortItem, expectItem) {
		t.Fatalf("expect %v, but get %v", expectItem, cachePortItem)
	}
}

// TestIsItemExisted test IsItemExisted function
func TestIsItemExisted(t *testing.T) {
	cache, err := getNewCache()
	if err != nil {
		t.Fatalf("failed to get new cache")
	}
	if !cache.IsItemExisted("test1.ns1", networkextensionv1.GetPoolItemKey("item1", []string{"lb1", "lb2"})) {
		t.Fatalf("pool test1.ns1, item item1-lb1,lb2 should be existed")
	}
	if cache.IsItemExisted("test1.ns1", "item1-lb1,lb22") {
		t.Fatalf("pool test1.ns1, item item1-lb1,lb2 should not be existed")
	}
}

// TestAllocateAllProtocolPortBinding test AllocateAllProtocolPortBinding function
func TestAllocateAllProtocolPortBinding(t *testing.T) {
	cache, err := getNewCache()
	if err != nil {
		t.Fatalf("failed to get new cache")
	}

	cache.SetPortBindingUsed("test1.ns1", networkextensionv1.GetPoolItemKey("item1", []string{"lb1", "lb2"}), "TCP", 30000, 0)
	_, mapItem, err := cache.AllocateAllProtocolPortBinding("test1.ns1")
	if err != nil {
		t.Fatalf("allocate port binding failed, err %s", err.Error())
	}
	expectMap := map[string]AllocatedPortItem{
		"TCP": {
			PoolKey:     "test1.ns1",
			PoolItemKey: networkextensionv1.GetPoolItemKey("item1", []string{"lb1", "lb2"}),
			Protocol:    "TCP",
			StartPort:   30001,
			EndPort:     0,
			IsUsed:      true,
		},
		"UDP": {
			PoolKey:     "test1.ns1",
			PoolItemKey: networkextensionv1.GetPoolItemKey("item1", []string{"lb1", "lb2"}),
			Protocol:    "UDP",
			StartPort:   30001,
			EndPort:     0,
			IsUsed:      true,
		},
	}
	if !reflect.DeepEqual(mapItem, expectMap) {
		t.Fatalf("expect %v, but get %v", mapItem, expectMap)
	}
}
