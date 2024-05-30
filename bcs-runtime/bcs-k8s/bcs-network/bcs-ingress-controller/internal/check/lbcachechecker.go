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

package check

import (
	"container/heap"
	"context"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// IngressChecker do ingress related check
type IngressChecker struct {
	cli      client.Client
	lbClient cloud.LoadBalance

	// cache for cloud loadbalance id
	lbIDCache *gocache.Cache
	// cache for cloud loadbalance name
	lbNameCache *gocache.Cache
	// expiration of lb cache, unit: minute
	lbCacheExpiration int
}

// NewIngressChecker return new ingress checker
func NewIngressChecker(cli client.Client, lbClient cloud.LoadBalance, lbIDCache, lbNameCache *gocache.Cache,
	lbCacheExpiration int) *IngressChecker {
	return &IngressChecker{
		cli:               cli,
		lbClient:          lbClient,
		lbIDCache:         lbIDCache,
		lbNameCache:       lbNameCache,
		lbCacheExpiration: lbCacheExpiration,
	}
}

// Run start check
func (ic *IngressChecker) Run() {
	blog.Infof("ingress checker begin")

	ingressList := &networkextensionv1.IngressList{}
	if err := ic.cli.List(context.TODO(), ingressList); err != nil {
		blog.Errorf("list ingress failed, err: %s", err.Error())
		return
	}

	// 初始化堆， 小顶堆顺序保存lb
	lbHeap := &minHeap{}
	heap.Init(lbHeap)
	viewedLB := sets.NewString()
	renewCnt := 0

	for _, ingress := range ingressList.Items {
		protocolLayer := common.GetIngressProtocolLayer(&ingress)
		for _, lbStatus := range ingress.Status.Loadbalancers {
			// 如果当前ingress的lb缓存消失或即将超时，从腾讯云重新拉信息并刷新缓存
			_, expiration, exist := ic.lbIDCache.GetWithExpiration(lbStatus.BuildKeyID())
			expireTime := 0
			if exist {
				expireTime = int(expiration.Sub(time.Now()).Minutes())
			}
			// 只排查缓存中有的LB， 避免频繁查询已被删除的LB
			if exist && !viewedLB.Has(lbStatus.BuildKeyID()) {
				viewedLB.Insert(lbStatus.BuildKeyID())
				info := lbInfo{
					LoadBalancerID:  lbStatus.LoadbalancerID,
					Region:          lbStatus.Region,
					EntityNamespace: ingress.Namespace,
					ProtocolLayer:   protocolLayer,
					ExpireTime:      expireTime,
				}

				if expireTime < 5 { // 过期时间少于5分钟强制刷新
					renewCnt++
					ic.renewCache(info)
				} else if expireTime < 30 { // 从过期时间少于30分钟的缓存中选出部分过期时间最短的刷新
					heap.Push(lbHeap, info)
				}
			}
		}
	}

	portpoolList := &networkextensionv1.PortPoolList{}
	if err := ic.cli.List(context.TODO(), portpoolList); err != nil {
		blog.Errorf("list portpool failed, err: %s", err.Error())
		return
	}
	for _, portpool := range portpoolList.Items {
		for _, itemStatus := range portpool.Status.PoolItemStatuses {
			for _, lbStatus := range itemStatus.PoolItemLoadBalancers {
				// 如果当前ingress的lb缓存消失或即将超时，从腾讯云重新拉信息并刷新缓存
				_, expiration, exist := ic.lbIDCache.GetWithExpiration(lbStatus.BuildKeyID())
				expireTime := 0
				if exist {
					expireTime = int(expiration.Sub(time.Now()).Minutes())
				}
				// 只排查缓存中有的LB， 避免频繁查询已被删除的LB
				if exist && !viewedLB.Has(lbStatus.BuildKeyID()) {
					viewedLB.Insert(lbStatus.BuildKeyID())
					info := lbInfo{
						LoadBalancerID:  lbStatus.LoadbalancerID,
						Region:          lbStatus.Region,
						EntityNamespace: portpool.Namespace,
						ProtocolLayer:   constant.ProtocolLayerTransport, // 端口池仅支持四层协议
						ExpireTime:      expireTime,
					}

					if expireTime < 5 { // 过期时间少于5分钟强制刷新
						renewCnt++
						ic.renewCache(info)
					} else if expireTime < 30 { // 从过期时间少于30分钟的缓存中选出部分过期时间最短的刷新
						heap.Push(lbHeap, info)
					}
				}
			}
		}
	}

	for i := 0; i < 15-renewCnt && lbHeap.Len() != 0; i++ {
		ic.renewCache(heap.Pop(lbHeap).(lbInfo))
	}

	blog.Infof("ingress checker done")
}

func (ic *IngressChecker) renewCache(info lbInfo) {
	var lbObj *cloud.LoadBalanceObject
	var err error

	if ic.lbClient.IsNamespaced() {
		lbObj, err = ic.lbClient.DescribeLoadBalancerWithNs(info.EntityNamespace, info.Region,
			info.LoadBalancerID, "", info.ProtocolLayer)
	} else {
		lbObj, err = ic.lbClient.DescribeLoadBalancer(info.Region, info.LoadBalancerID, "",
			info.ProtocolLayer)
	}

	if err != nil {
		if err == cloud.ErrLoadbalancerNotFound {
			// 删除对应缓存，避免后续重复尝试刷新
			blog.Warnf("%s/%s lb '%s' not found", info.EntityNamespace, info.EntityName, info.LoadBalancerID)
			ic.lbIDCache.Delete(info.Region + ":" + info.LoadBalancerID)
			ic.lbNameCache.Delete(info.Region + ":" + info.LoadBalancerName)
		} else {
			blog.Errorf("describe loadbalancer '%s' from cloud failed, ingress: %s/%s, err: %s",
				info.LoadBalancerID, info.EntityNamespace, info.EntityName, err.Error())
		}
		return
	}

	ic.lbIDCache.SetDefault(info.Region+":"+info.LoadBalancerID, lbObj)
	ic.lbNameCache.SetDefault(info.Region+":"+info.LoadBalancerName, lbObj)
}

type lbInfo struct {
	LoadBalancerID   string
	LoadBalancerName string
	Region           string
	EntityNamespace  string
	EntityName       string
	ProtocolLayer    string

	ExpireTime int // 过期剩余时间，单位分钟
}

type minHeap []lbInfo

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].ExpireTime < h[j].ExpireTime }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *minHeap) Push(x interface{}) {
	*h = append(*h, x.(lbInfo))
}

func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	if n <= 0 {
		return nil
	}
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
