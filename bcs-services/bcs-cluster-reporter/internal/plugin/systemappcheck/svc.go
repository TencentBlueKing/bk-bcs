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

// Package systemappcheck xxx
package systemappcheck

import (
	"fmt"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metricmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// CheckService xxx
func CheckService(cluster *pluginmanager.ClusterConfig, clusterID string) ([]pluginmanager.CheckItem, []*metricmanager.GaugeVecSet, error) {
	namespaceList, err := cluster.ClientSet.CoreV1().Namespaces().List(util.GetCtx(time.Second*10), metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		return nil, nil, err
	}

	checkItemList := make([]pluginmanager.CheckItem, 0, 0)
	gvsList := make([]*metricmanager.GaugeVecSet, 0, 0)

	var wg sync.WaitGroup
	routinePool := util.NewRoutinePool(20)
	for _, namespace := range namespaceList.Items {
		wg.Add(1)
		go func(namespace v1.Namespace) {
			routinePool.Add(1)
			defer func() {
				wg.Done()
				routinePool.Done()
			}()

			serviceList, err := cluster.ClientSet.CoreV1().Services(namespace.Name).List(util.GetCtx(time.Second*10), metav1.ListOptions{ResourceVersion: "0"})
			if err != nil {
				klog.Errorf("%s get service in namespace %s failed: %s", clusterID, namespace.Name, err.Error())
				return
			}

			for _, svc := range serviceList.Items {
				if svc.Spec.Type == "LoadBalancer" {
					if len(svc.Status.LoadBalancer.Ingress) == 0 {
						checkItemList = append(checkItemList, pluginmanager.CheckItem{
							ItemName:   SystemAppConfigCheckItem,
							ItemTarget: svc.Name,
							Status:     ConfigErrorStatus,
							Normal:     false,
							Detail:     fmt.Sprintf(StringMap[lbSVCNoIpDetail], svc.Namespace, svc.Name),
							Tags:       nil,
							Level:      pluginmanager.WARNLevel,
						})
						gvsList = append(gvsList, &metricmanager.GaugeVecSet{
							Labels: []string{cluster.ClusterID, cluster.BusinessID, namespace.Name, svc.Name, "service", ConfigErrorStatus},
							Value:  1,
						})
					}
				}
			}
		}(namespace)

	}
	wg.Wait()

	return checkItemList, gvsList, nil
}
