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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metricmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
	"k8s.io/klog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CheckStaticPod check static pod config
func CheckStaticPod(cluster *pluginmanager.ClusterConfig) ([]pluginmanager.CheckItem, []*metricmanager.GaugeVecSet, error) {
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

	checkItemList := make([]pluginmanager.CheckItem, 0, 0)
	gvsList := make([]*metricmanager.GaugeVecSet, 0, 0)

	newStaticPodNameList := make([]string, 0, 0)

	etcdPodList := make([]*v1.Pod, 0, 0)
	apiserverPodList := make([]*v1.Pod, 0, 0)
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
			apiserverPodList = append(apiserverPodList, &pod)
		}
		if strings.HasPrefix(pod.Name, "etcd") {
			etcdPodList = append(etcdPodList, &pod)
		}

		if strings.HasPrefix(pod.Name, "kube-controller-manager") {
			checkItemList = append(checkItemList, CheckKCM(&pod, cluster)...)
		}

		checkItemList = append(checkItemList, CheckLabel(&pod)...)
	}
	checkItemList = append(checkItemList, CheckETCD(etcdPodList, cluster)...)
	checkItemList = append(checkItemList, CheckApiserver(apiserverPodList, cluster)...)

	if !ok {
		// 缓存集群当前static pod信息，避免频繁list pod
		util.SetCacheWithTimeout(cluster.ClusterID+"staticpod", newStaticPodNameList, time.Hour*24)
	}

	if len(checkItemList) == 0 {
		checkItemList = append(checkItemList, pluginmanager.CheckItem{
			ItemName:   SystemAppConfigCheckItem,
			ItemTarget: "static pod",
			Normal:     true,
			Detail:     "",
			Tags:       nil,
			Level:      pluginmanager.WARNLevel,
		})
	}

	// 生成异常配置指标
	okFlag := true
	for index, checkItem := range checkItemList {
		if !checkItem.Normal {
			gvsList = append(gvsList, &metricmanager.GaugeVecSet{
				Labels: []string{cluster.ClusterID, cluster.BusinessID, "kube-system", checkItem.ItemTarget, "pod", checkItem.Status},
				Value:  1,
			})
			okFlag = false
		}
		checkItemList[index] = checkItem
	}

	if okFlag {
		gvsList = append(gvsList, &metricmanager.GaugeVecSet{
			Labels: []string{cluster.ClusterID, cluster.BusinessID, "kube-system", "static pod", "pod", pluginmanager.NormalStatus},
			Value:  1,
		})
	}

	// 静态pod参数不一致问题检查
	return checkItemList, gvsList, nil
}

// CheckKCM check kcm config
func CheckKCM(pod *v1.Pod, cluster *pluginmanager.ClusterConfig) []pluginmanager.CheckItem {
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

// CheckETCD check etcd config
func CheckETCD(podList []*v1.Pod, cluster *pluginmanager.ClusterConfig) []pluginmanager.CheckItem {
	checkItemList := make([]pluginmanager.CheckItem, 0)
	podParamList := make([][]string, 0)

	// 检查参数
	floatFlagList := []plugin.FloatFlag{
		{Name: "--heartbeat-interval",
			CompareType: "ge",
			Value:       1000,
			Needed:      true,
		},
	}

	for _, pod := range podList {
		for _, floatFlag := range floatFlagList {
			detail := plugin.CheckFlag(append(pod.Spec.Containers[0].Command, pod.Spec.Containers[0].Args...), floatFlag)
			if detail != "" {
				checkItemList = append(checkItemList, pluginmanager.CheckItem{
					ItemName:   SystemAppConfigCheckItem,
					ItemTarget: pod.Name,
					Status:     UnrecommandedStatus,
					Normal:     false,
					Detail:     detail,
					Tags:       nil,
					Level:      pluginmanager.WARNLevel,
				})
			}
		}

		// 磁盘配置
		checkFlag := false
		for _, volume := range pod.Spec.Volumes {
			if volume.HostPath.Path == "/var/lib/etcd" {
				checkFlag = true
				checkItemList = append(checkItemList, pluginmanager.CheckItem{
					ItemName:   SystemAppConfigCheckItem,
					ItemTarget: pod.Name,
					Status:     UnrecommandedStatus,
					Normal:     false,
					Detail:     fmt.Sprintf(StringMap[etcdDataDiskDetail], volume.HostPath.Path),
					Tags:       nil,
					Level:      pluginmanager.WARNLevel,
				})
				break
			}
		}

		if !checkFlag {
			checkItemList = append(checkItemList, pluginmanager.CheckItem{
				ItemName:   SystemAppConfigCheckItem,
				ItemTarget: "etcd",
				Status:     pluginmanager.NormalStatus,
				Normal:     true,
				Detail:     "",
				Tags:       nil,
				Level:      pluginmanager.WARNLevel,
			})
		}

		podParamList = append(podParamList, pod.Spec.Containers[0].Args)
	}

	// 检查etcd参数是否一致
	err := checkParamConsistency(podParamList, nil)
	if err != nil {
		klog.Errorf("%s checkParamConsistency failed: %s", cluster.ClusterID, err.Error())
		checkItemList = append(checkItemList, pluginmanager.CheckItem{
			ItemName:   SystemAppConfigCheckItem,
			ItemTarget: "etcd",
			Status:     ConfigInconsistencyStatus,
			Normal:     false,
			Detail:     err.Error(),
			Tags:       nil,
			Level:      pluginmanager.WARNLevel,
		})
	}

	// 检查状态
	return checkItemList
}

// CheckApiserver check apiserver config
func CheckApiserver(podList []*v1.Pod, cluster *pluginmanager.ClusterConfig) []pluginmanager.CheckItem {
	checkItemList := make([]pluginmanager.CheckItem, 0, 0)
	podParamList := make([][]string, 0)

	for _, pod := range podList {
		podParamList = append(podParamList, append(pod.Spec.Containers[0].Command, pod.Spec.Containers[0].Args...))
	}

	// 检查etcd参数是否一致
	err := checkParamConsistency(podParamList, []string{"--goaway-chance", "--audit-policy-file"})
	if err != nil {
		klog.Errorf("%s checkParamConsistency failed: %s", cluster.ClusterID, err.Error())
		checkItemList = append(checkItemList, pluginmanager.CheckItem{
			ItemName:   SystemAppConfigCheckItem,
			ItemTarget: "etcd",
			Status:     ConfigInconsistencyStatus,
			Normal:     false,
			Detail:     err.Error(),
			Tags:       nil,
			Level:      pluginmanager.WARNLevel,
		})
	}

	return checkItemList
}

// CheckLabel xxx
func CheckLabel(pod *v1.Pod) []pluginmanager.CheckItem {
	checkItem := pluginmanager.CheckItem{
		ItemName:   SystemAppConfigCheckItem,
		ItemTarget: pod.Name,
		Detail:     fmt.Sprintf(StringMap[NoLabelDetailFormat], pod.Name),
		Tags:       nil,
		Level:      pluginmanager.RISKLevel,
	}

	result := make([]pluginmanager.CheckItem, 0, 0)
	if len(pod.Labels) == 0 {
		checkItem.Status = NolabelStatus
		checkItem.Normal = false
	} else {
		checkItem.Status = pluginmanager.NormalStatus
		checkItem.Normal = true
		checkItem.Detail = ""
	}

	result = append(result, checkItem)

	return result
}

// CheckSystemWorkLoadConfig 检查系统应用配置
func CheckSystemWorkLoadConfig(cluster *pluginmanager.ClusterConfig) ([]pluginmanager.CheckItem, []*metricmanager.GaugeVecSet) {
	checkItemList := make([]pluginmanager.CheckItem, 0, 0)
	checkItemList = append(checkItemList, CheckCoredns(cluster.ClientSet)...)
	checkItemList = append(checkItemList, CheckKubeProxy(cluster.ClientSet)...)

	gvsList := make([]*metricmanager.GaugeVecSet, 0, 0)

	for _, checkItem := range checkItemList {
		gvsList = append(gvsList, &metricmanager.GaugeVecSet{
			Labels: []string{cluster.ClusterID, cluster.BusinessID, "kube-system", checkItem.ItemTarget, "app", checkItem.Status},
			Value:  1,
		})
	}

	return checkItemList, gvsList
}

// CheckCoredns 检查coredns config
func CheckCoredns(clientSet *kubernetes.Clientset) []pluginmanager.CheckItem {
	result := make([]pluginmanager.CheckItem, 0, 0)
	checkItem := pluginmanager.CheckItem{
		ItemName:   SystemAppConfigCheckItem,
		ItemTarget: "coredns",
		Tags:       nil,
		Level:      pluginmanager.RISKLevel,
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

	// 检查coredns是否配置了健康检查端口以及lameduck配置
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
		checkItem.Status = pluginmanager.NormalStatus
		checkItem.Normal = true
		result = append(result, checkItem)
	}

	return result
}

// CheckKubeProxy check kube-proxy config
func CheckKubeProxy(clientSet *kubernetes.Clientset) []pluginmanager.CheckItem {
	result := make([]pluginmanager.CheckItem, 0, 0)
	checkItem := pluginmanager.CheckItem{
		ItemName:   SystemAppConfigCheckItem,
		ItemTarget: "kube-proxy",
		Tags:       nil,
		Level:      pluginmanager.RISKLevel,
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

	// 检查proxy模式是否为ipvs以及udp timeout是否设置
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
		checkItem.Status = pluginmanager.NormalStatus
	}

	result = append(result, checkItem)
	return result
}

func checkParamConsistency(podsParamList [][]string, mustContain []string) error {
	if len(podsParamList) < 1 {
		return nil
	}

	paramMap := make(map[string]string)

	for _, param := range podsParamList[0] {
		// 只检查键值对类型的参数
		if !strings.Contains(param, "=") {
			return nil
		}

		if strings.HasPrefix(param, "--") {
			param = strings.TrimPrefix(param, "--")
		}

		paramName, paramValue := strings.SplitN(param, "=", 2)[0], strings.SplitN(param, "=", 2)[1]
		paramMap[paramName] = paramValue
	}

	podsParamList = podsParamList[1:]

	for _, paramList := range podsParamList {
		for _, param := range paramList {
			// 只检查键值对类型的参数
			if !strings.Contains(param, "=") {
				return nil
			}

			if strings.HasPrefix(param, "--") {
				param = strings.TrimPrefix(param, "--")
			}

			// 不校验包含IP的参数
			if containsIP(param) {
				continue
			}

			paramName, paramValue := strings.SplitN(param, "=", 2)[0], strings.SplitN(param, "=", 2)[1]
			if value, ok := paramMap[paramName]; !ok {
				return fmt.Errorf("check param %s doesn't exist in all pod, inconsistency", paramName)
			} else if value != paramValue {
				return fmt.Errorf("check param %s is %s and %s, inconsistency", paramName, value, paramValue)
			}
		}
	}

	if mustContain != nil {
		for _, param := range mustContain {
			if _, ok := paramMap[param]; !ok {
				return fmt.Errorf("check param %s doesn't exist", param)
			}
		}
	}
	return nil
}

func containsIP(s string) bool {
	ipRegex := `\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`
	match, _ := regexp.MatchString(ipRegex, s)
	return match
}
