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
 *
 */

package pod_monitor

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api/servicemonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

// ListPodMonitors 获取PodMonitor列表数据
// @Summary PodMonitor列表数据
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/:namespace/pod_monitors [get]
func ListPodMonitors(c *rest.Context) (interface{}, error) {
	podMonitor := make([]*v1.PodMonitor, 0)
	// 共享集群不展示列表
	if c.SharedCluster {
		return podMonitor, nil
	}
	limit := c.Query("limit")
	offset := c.Query("offset")
	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = c.Param("namespace")
	}
	client, err := servicemonitor.GetMonitoringV1Client(c)
	if err != nil {
		return nil, err
	}
	var limitInt int
	if limit != "" {
		limitInt, err = strconv.Atoi(limit)
		if err != nil {
			return nil, err
		}
	}
	listOps := metav1.ListOptions{
		Limit:    int64(limitInt),
		Continue: offset,
	}
	data, err := client.PodMonitors(namespace).List(c.Context, listOps)
	if err != nil {
		return nil, err
	}
	for _, v := range data.Items {
		podMonitor = append(podMonitor, v)
	}
	return podMonitor, nil
}

// CreatePodMonitor 创建PodMonitor
// @Summary PodMonitor列表数据
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/:namespace/pod_monitors [post]
func CreatePodMonitor(c *rest.Context) (interface{}, error) {
	podMonitorReq := &CreatePodMonitorReq{}
	if err := c.ShouldBindJSON(podMonitorReq); err != nil {
		return nil, err
	}
	podMonitorReq.Namespace = c.Param("namespace")

	flag := podMonitorReq.Validate()

	if !flag {
		return nil, errors.New("参数校验失败")
	}

	client, err := servicemonitor.GetMonitoringV1Client(c)
	if err != nil {
		return nil, err
	}
	params := make(map[string][]string, 0)
	for k, v := range podMonitorReq.Params {
		params[k] = []string{v}
	}

	podMonitor := &v1.PodMonitor{}
	labels := map[string]string{
		"release":                     "po",
		"io.tencent.paas.source_type": "bcs",
		"io.tencent.bcs.service_name": podMonitorReq.ServiceName,
	}

	podMonitor.ObjectMeta = metav1.ObjectMeta{
		Labels:    labels,
		Name:      podMonitorReq.Name,
		Namespace: podMonitorReq.Namespace,
	}
	endpoints := make([]v1.PodMetricsEndpoint, 0)
	initEndpoint := v1.PodMetricsEndpoint{
		Port:     podMonitorReq.Port,
		Path:     podMonitorReq.Path,
		Interval: podMonitorReq.Interval,
		Params:   params,
	}

	endpoints = append(endpoints, initEndpoint)
	podMonitor.Spec = v1.PodMonitorSpec{
		PodMetricsEndpoints: endpoints,
		SampleLimit:         uint64(podMonitorReq.SampleLimit),
		Selector: metav1.LabelSelector{
			MatchLabels: podMonitorReq.Selector,
		},
	}

	_, err = client.PodMonitors(podMonitorReq.Namespace).Create(c.Context, podMonitor, metav1.CreateOptions{})

	if err != nil {
		return nil, err
	}
	return nil, nil
}

// DeletePodMonitor 删除PodMonitor
// @Summary 删除PodMonitor
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/:namespace/pod_monitors/:name [delete]
func DeletePodMonitor(c *rest.Context) (interface{}, error) {
	name := c.Param("name")
	flag := validateName(name)
	if !flag {
		return nil, fmt.Errorf("校验name参数: %s 是否符合k8s资源名称格式并且长度不大于63位字符不通过", name)
	}
	namespace := c.Param("namespace")
	client, err := servicemonitor.GetMonitoringV1Client(c)
	if err != nil {
		return nil, err
	}
	err = client.PodMonitors(namespace).Delete(c, name, metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("删除 Metrics: %s 成功", name), nil
}

// BatchDeletePodMonitor 批量删除PodMonitor
// @Summary 批量删除PodMonitor
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /pod_monitors/batchdelete [post]
func BatchDeletePodMonitor(c *rest.Context) (interface{}, error) {
	// 共享集群禁止批量删除
	if c.SharedCluster {
		return nil, fmt.Errorf("denied")
	}
	podMonitorDelReq := &BatchDeletePodMonitorReq{}
	if err := c.ShouldBindJSON(podMonitorDelReq); err != nil {
		return nil, err
	}
	client, err := servicemonitor.GetMonitoringV1Client(c)
	if err != nil {
		return nil, err
	}
	for _, v := range podMonitorDelReq.PodMonitors {
		err = client.PodMonitors(v.Namespace).Delete(c, v.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, err
		}
	}

	return fmt.Sprintf("批量删除 Metrics: %s 成功", podMonitorDelReq), nil
}

// GetPodMonitor 获取单个PodMonitor
// @Summary 删除PodMonitor
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/:namespace/pod_monitors/:name [get]
func GetPodMonitor(c *rest.Context) (interface{}, error) {
	name := c.Param("name")
	flag := validateName(name)
	if !flag {
		return nil, fmt.Errorf("校验name参数: %s 是否符合k8s资源名称格式并且长度不大于63位字符不通过", name)
	}
	namespace := c.Param("namespace")
	client, err := servicemonitor.GetMonitoringV1Client(c)
	if err != nil {
		return nil, err
	}
	result, err := client.PodMonitors(namespace).Get(c.Context, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdatePodMonitor 创建PodMonitor
// @Summary PodMonitor列表数据
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/:namespace/pod_monitors/:name [put]
func UpdatePodMonitor(c *rest.Context) (interface{}, error) {
	podMonitorReq := &UpdatePodMonitorReq{}
	if err := c.ShouldBindJSON(podMonitorReq); err != nil {
		return nil, err
	}
	podMonitorReq.Name = c.Param("name")
	podMonitorReq.Namespace = c.Param("namespace")

	flag := podMonitorReq.Validate()

	if !flag {
		return nil, errors.New("参数校验失败")
	}

	client, err := servicemonitor.GetMonitoringV1Client(c)
	if err != nil {
		return nil, err
	}

	exist, err := client.PodMonitors(podMonitorReq.Namespace).
		Get(c.Context, podMonitorReq.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	podMonitor := exist.DeepCopy()
	labels := exist.Labels
	labels["release"] = "po"
	labels["io.tencent.paas.source_type"] = "bcs"
	labels["io.tencent.bcs.service_name"] = podMonitorReq.ServiceName

	params := make(map[string][]string, 0)
	for k, v := range podMonitorReq.Params {
		params[k] = []string{v}
	}
	podMonitor.Labels = labels
	endpoints := make([]v1.PodMetricsEndpoint, 0)
	initEndpoint := v1.PodMetricsEndpoint{
		Port:     podMonitorReq.Port,
		Path:     podMonitorReq.Path,
		Interval: podMonitorReq.Interval,
		Params:   params,
	}

	endpoints = append(endpoints, initEndpoint)
	podMonitor.Spec = v1.PodMonitorSpec{
		PodMetricsEndpoints: endpoints,
		SampleLimit:         uint64(podMonitorReq.SampleLimit),
		Selector: metav1.LabelSelector{
			MatchLabels: podMonitorReq.Selector,
		},
	}
	podMonitorClient := client.PodMonitors(podMonitorReq.Namespace)
	_, err = podMonitorClient.Update(c.Context, podMonitor, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// validateName 校验name参数是否符合k8s资源名称格式并且长度不大于63位字符
func validateName(name string) bool {
	if len(name) > 63 {
		return false
	}
	if match, _ := regexp.MatchString("^[a-z][-a-z0-9]*$", name); !match {
		return false
	}

	return true
}

// validatePath 校验参数是否合法，不可为空
func validateSelector(selector map[string]string) bool {
	if selector == nil || len(selector) == 0 {
		return false
	}
	return true
}

// validateSampleLimit 校验参数是否合法
func validateSampleLimit(samplelimit int) bool {
	if SampleLimitMax >= samplelimit && samplelimit >= SampleLimitMin {
		return true
	}
	return false
}
