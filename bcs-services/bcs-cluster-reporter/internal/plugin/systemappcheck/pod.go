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
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// TODO 静态pod参数不一致问题检查
// CheckStaticPod
func CheckStaticPod(cluster *plugin_manager.ClusterConfig) ([]plugin_manager.CheckItem, []*metric_manager.GaugeVecSet, error) {
	staticPodcache, ok := util.GetCache(cluster.ClusterID + "staticpod")
	podList := make([]v1.Pod, 0, 0)
	if ok {
		staticPodNameList, ok1 := staticPodcache.([]string)
		klog.Infof("%s has static pod caches, get pod from kube-system namespace", cluster.ClusterID)
		if !ok1 {
			return nil, nil, fmt.Errorf("%s get staticPodcache failed %s", cluster.ClusterID, staticPodcache)
		}

		for _, staticPodName := range staticPodNameList {
			pod, err := cluster.ClientSet.CoreV1().Pods("kube-system").Get(context.Background(), staticPodName, metav1.GetOptions{ResourceVersion: "0"})
			if err != nil {
				if errors.IsNotFound(err) {
					ok = false
				}
				klog.Errorf("%s get static pod failed: %s", cluster.ClusterID, err.Error())
			} else {
				podList = append(podList, *pod)
			}
		}
	}
	if !ok {
		klog.Infof("%s has no static pod caches, list from cluster kube-system namespace", cluster.ClusterID)
		result, err := cluster.ClientSet.CoreV1().Pods("kube-system").List(context.Background(), metav1.ListOptions{ResourceVersion: "0"})
		if err != nil {
			return nil, nil, err
		} else {
			podList = result.Items
		}
	}

	checkItemList := make([]plugin_manager.CheckItem, 0, 0)
	gvsList := make([]*metric_manager.GaugeVecSet, 0, 0)

	newStaticPodNameList := make([]string, 0, 0)
	for _, pod := range podList {
		if len(pod.OwnerReferences) == 0 {
			continue
		}
		if pod.OwnerReferences[0].Kind != "Node" {
			continue
		}

		newStaticPodNameList = append(newStaticPodNameList, pod.Name)

		if strings.HasPrefix(pod.Name, "cloud-controller-manager") {
			continue
		}

		if strings.HasPrefix(pod.Name, "kube-apiserver") {
			checkItemList = append(checkItemList, CheckApiserver(&pod, cluster)...)
		}
		if strings.HasPrefix(pod.Name, "etcd") {
			checkItemList = append(checkItemList, CheckETCD(&pod, cluster)...)
		}
		if strings.HasPrefix(pod.Name, "kube-controller-manager") {
			checkItemList = append(checkItemList, CheckKCM(&pod, cluster)...)
		}

		checkItemList = append(checkItemList, CheckLabel(&pod)...)
	}

	if !ok {
		util.SetCacheWithTimeout(cluster.ClusterID+"staticpod", newStaticPodNameList, time.Hour*24)
	}

	if len(checkItemList) == 0 {
		checkItemList = append(checkItemList, plugin_manager.CheckItem{
			ItemName:   SystemAppConfigCheckItem,
			ItemTarget: "static pod",
			Normal:     true,
			Detail:     "",
			Tags:       nil,
			Level:      plugin_manager.WARNLevel,
		})
	}

	// 生成异常配置指标
	okFlag := true
	for index, checkItem := range checkItemList {
		if !checkItem.Normal {
			gvsList = append(gvsList, &metric_manager.GaugeVecSet{
				Labels: []string{cluster.ClusterID, cluster.BusinessID, "kube-system", checkItem.ItemTarget, "pod", checkItem.Status},
				Value:  1,
			})
			okFlag = false
		}
		checkItemList[index] = checkItem
	}

	if okFlag {
		gvsList = append(gvsList, &metric_manager.GaugeVecSet{
			Labels: []string{cluster.ClusterID, cluster.BusinessID, "kube-system", "static pod", "pod", plugin_manager.NormalStatus},
			Value:  1,
		})
	}

	return checkItemList, gvsList, nil
}

func CheckKCM(pod *v1.Pod, cluster *plugin_manager.ClusterConfig) []plugin_manager.CheckItem {
	cidrFlag := false
	cidr := make([]string, 0, 0)
	for _, arg := range append(pod.Spec.Containers[0].Command, pod.Spec.Containers[0].Args...) {
		if strings.HasPrefix(arg, "--service-cluster-ip-range") {
			cluster.ServiceCidr = strings.SplitN(arg, "=", 2)[1]
		} else if strings.HasPrefix(arg, "--cluster-cidr") {
			cidr = append(cidr, strings.SplitN(arg, "=", 2)[1])
		} else if arg == "--allocate-node-cidrs=true" {
			cidrFlag = true
		} else if strings.HasPrefix(arg, "--node-cidr-mask-size") {
			cluster.MaskSize, _ = strconv.Atoi(strings.SplitN(arg, "=", 2)[1])
		}
	}

	// 重置cidr
	cluster.Cidr = make([]string, 0, 0)
	if cidrFlag {
		cluster.Cidr = append(cluster.Cidr, cidr...)
	}

	if cluster.MaskSize == 0 {
		cluster.MaskSize = 24
	}

	return nil
}

func CheckETCD(pod *v1.Pod, cluster *plugin_manager.ClusterConfig) []plugin_manager.CheckItem {
	checkItemList := make([]plugin_manager.CheckItem, 0, 0)

	// 检查参数
	floatFlagList := []plugin.FloatFlag{
		{Name: "--heartbeat-interval",
			CompareType: "ge",
			Value:       1000,
			Needed:      true,
		},
	}

	for _, floatFlag := range floatFlagList {
		detail := plugin.CheckFlag(append(pod.Spec.Containers[0].Command, pod.Spec.Containers[0].Args...), floatFlag)
		if detail != "" {
			checkItemList = append(checkItemList, plugin_manager.CheckItem{
				ItemName:   SystemAppConfigCheckItem,
				ItemTarget: pod.Name,
				Status:     UnrecommandedStatus,
				Normal:     false,
				Detail:     detail,
				Tags:       nil,
				Level:      plugin_manager.WARNLevel,
			})
		}
	}

	// 磁盘配置
	checkFlag := false
	for _, volume := range pod.Spec.Volumes {
		if volume.HostPath.Path == "/var/lib/etcd" {
			checkFlag = true
			checkItemList = append(checkItemList, plugin_manager.CheckItem{
				ItemName:   SystemAppConfigCheckItem,
				ItemTarget: pod.Name,
				Status:     ConfigErrorStatus,
				Normal:     false,
				Detail:     fmt.Sprintf(StringMap[etcdDataDiskDetail], volume.HostPath.Path),
				Tags:       nil,
				Level:      plugin_manager.WARNLevel,
			})
			break
		}
	}

	if !checkFlag {
		checkItemList = append(checkItemList, plugin_manager.CheckItem{
			ItemName:   SystemAppConfigCheckItem,
			ItemTarget: pod.Name,
			Status:     plugin_manager.NormalStatus,
			Normal:     true,
			Detail:     "",
			Tags:       nil,
			Level:      plugin_manager.WARNLevel,
		})
	}

	// 检查状态
	return checkItemList
}

func CheckApiserver(pod *v1.Pod, cluster *plugin_manager.ClusterConfig) []plugin_manager.CheckItem {
	checkItemList := make([]plugin_manager.CheckItem, 0, 0)

	setFlagList := []string{"--goaway-chance", "--audit-policy-file"}

	for _, arg := range append(pod.Spec.Containers[0].Command, pod.Spec.Containers[0].Args...) {
		for index, flag := range setFlagList {
			if strings.Contains(arg, flag) && flag != "" {
				checkItemList = append(checkItemList, plugin_manager.CheckItem{
					ItemName:   SystemAppConfigCheckItem,
					ItemTarget: pod.Name,
					Status:     plugin_manager.NormalStatus,
					Normal:     true,
					Detail:     "",
					Tags:       nil,
					Level:      plugin_manager.WARNLevel,
				})
				setFlagList[index] = ""
			}
		}

		if strings.HasPrefix(arg, "--service-cluster-ip-range") {
			cluster.ServiceCidr = strings.SplitN(arg, "=", 2)[1]
		}
	}

	for _, setFlag := range setFlagList {
		if setFlag != "" {
			checkItemList = append(checkItemList, plugin_manager.CheckItem{
				ItemName:   SystemAppConfigCheckItem,
				ItemTarget: pod.Name,
				Status:     ConfigNotFoundStatus,
				Normal:     false,
				Detail:     fmt.Sprintf(StringMap[FlagUnsetDetailFormat], setFlag),
				Tags:       nil,
				Level:      plugin_manager.WARNLevel,
			})
			return checkItemList
		}
	}

	return checkItemList

	// 检查参数
	// 检查状态
}

func CheckLabel(pod *v1.Pod) []plugin_manager.CheckItem {
	checkItem := plugin_manager.CheckItem{
		ItemName:   SystemAppConfigCheckItem,
		ItemTarget: pod.Name,
		Detail:     fmt.Sprintf(StringMap[NoLabelDetailFormat], pod.Name),
		Tags:       nil,
		Level:      plugin_manager.RISKLevel,
	}

	result := make([]plugin_manager.CheckItem, 0, 0)
	if len(pod.Labels) == 0 {
		checkItem.Status = NolabelStatus
		checkItem.Normal = false
	} else {
		checkItem.Status = plugin_manager.NormalStatus
		checkItem.Normal = true
	}

	result = append(result, checkItem)

	return result
}

func CheckSystemWorkLoadConfig(clientSet *kubernetes.Clientset) []plugin_manager.CheckItem {
	result := make([]plugin_manager.CheckItem, 0, 0)
	result = append(result, CheckCoredns(clientSet)...)
	result = append(result, CheckKubeProxy(clientSet)...)

	return result
}

// TODO 代码逻辑优化
func CheckCoredns(clientSet *kubernetes.Clientset) []plugin_manager.CheckItem {
	result := make([]plugin_manager.CheckItem, 0, 0)
	checkItem := plugin_manager.CheckItem{
		ItemName:   SystemAppConfigCheckItem,
		ItemTarget: "coredns",
		Tags:       nil,
		Level:      plugin_manager.RISKLevel,
	}

	cm, err := clientSet.CoreV1().ConfigMaps("kube-system").Get(util.GetCtx(10*time.Second), "coredns", metav1.GetOptions{ResourceVersion: "0"})

	if err != nil {
		checkItem.Normal = false
		if strings.Contains(err.Error(), "not found") {
			checkItem.Status = ConfigNotFoundStatus
		} else {
			checkItem.Status = ConfigErrorStatus
		}
		checkItem.Detail = fmt.Sprintf(StringMap[GetResourceFailedDetail], "coredns configmap", err.Error())
		result = append(result, checkItem)
		return result
	}

	flagList := []string{
		"ready", "lameduck",
	}

	unSetFlagList := make([]string, 0, 0)
	for _, flag := range flagList {
		if !strings.Contains(cm.Data["Corefile"], flag) {
			checkItem.Detail = fmt.Sprintf(StringMap[FlagUnsetDetailFormat], unSetFlagList)
			checkItem.Normal = false
			checkItem.Status = ConfigErrorStatus
			result = append(result, checkItem)
			return result
		}
	}

	if len(result) == 0 {
		checkItem.Status = plugin_manager.NormalStatus
		checkItem.Normal = true
		result = append(result, checkItem)
	}

	return result
}

func CheckKubeProxy(clientSet *kubernetes.Clientset) []plugin_manager.CheckItem {
	result := make([]plugin_manager.CheckItem, 0, 0)
	checkItem := plugin_manager.CheckItem{
		ItemName:   SystemAppConfigCheckItem,
		ItemTarget: "kube-proxy",
		Tags:       nil,
		Level:      plugin_manager.RISKLevel,
		Normal:     true,
	}

	ds, err := clientSet.AppsV1().DaemonSets("kube-system").Get(util.GetCtx(10*time.Second), "kube-proxy", metav1.GetOptions{ResourceVersion: "0"})

	if err != nil {
		checkItem.Normal = false
		if strings.Contains(err.Error(), "not found") {
			checkItem.Status = ConfigNotFoundStatus
		} else {
			checkItem.Status = ConfigErrorStatus
		}
		checkItem.Detail = err.Error()
		result = append(result, checkItem)
		return result
	}

	var ipvsFlag, udpTimeoutFlag bool
	for _, arg := range append(ds.Spec.Template.Spec.Containers[0].Command, ds.Spec.Template.Spec.Containers[0].Args...) {
		if strings.Contains(arg, "proxy-mode=ipvs") {
			ipvsFlag = true
		} else if strings.Contains(arg, "ipvs-udp-timeout=10s") {
			udpTimeoutFlag = true
		}
	}

	if ipvsFlag && !udpTimeoutFlag {
		checkItem.Normal = false
		checkItem.Detail = StringMap[kubeProxyIpvsDetail]
		checkItem.Status = ConfigErrorStatus
	} else {
		checkItem.Status = plugin_manager.NormalStatus
	}

	result = append(result, checkItem)
	return result
}
