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

// Package podmonitor prometheus pods monitor
package podmonitor

import (
	"context"
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
func ListPodMonitors(c context.Context, req *ListPodMonitorsReq) (*[]*v1.PodMonitor, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	podMonitor := make([]*v1.PodMonitor, 0)
	// 共享集群不展示列表
	if rctx.SharedCluster {
		return &podMonitor, nil
	}
	limit := req.Limit
	offset := req.Offset
	namespace := req.Namespace
	if namespace == "" {
		namespace = req.PathNamespace
	}
	client, err := servicemonitor.GetMonitoringV1Client(rctx)
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
	data, err := client.PodMonitors(namespace).List(c, listOps)
	if err != nil {
		return nil, err
	}
	podMonitor = append(podMonitor, data.Items...)
	return &podMonitor, nil
}

// CreatePodMonitor 创建PodMonitor
// @Summary PodMonitor列表数据
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/:namespace/pod_monitors [post]
func CreatePodMonitor(c context.Context, req *CreatePodMonitorReq) (*any, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	req.Namespace = req.PathNamespace

	flag := req.Validate()

	if !flag {
		return nil, errors.New("参数校验失败")
	}

	client, err := servicemonitor.GetMonitoringV1Client(rctx)
	if err != nil {
		return nil, err
	}
	params := make(map[string][]string, 0)
	for k, v := range req.Params {
		params[k] = []string{v}
	}

	podMonitor := &v1.PodMonitor{}
	labels := map[string]string{
		"release":                     "po",
		"io.tencent.paas.source_type": "bcs",
		"io.tencent.bcs.service_name": req.ServiceName,
	}

	podMonitor.ObjectMeta = metav1.ObjectMeta{
		Labels:    labels,
		Name:      req.Name,
		Namespace: req.Namespace,
	}
	endpoints := make([]v1.PodMetricsEndpoint, 0)
	initEndpoint := v1.PodMetricsEndpoint{
		Port:     req.Port,
		Path:     req.Path,
		Interval: req.Interval,
		Params:   params,
	}

	endpoints = append(endpoints, initEndpoint)
	podMonitor.Spec = v1.PodMonitorSpec{
		PodMetricsEndpoints: endpoints,
		SampleLimit:         uint64(req.SampleLimit),
		Selector: metav1.LabelSelector{
			MatchLabels: req.Selector,
		},
	}

	_, err = client.PodMonitors(req.Namespace).Create(c, podMonitor, metav1.CreateOptions{})

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
func DeletePodMonitor(c context.Context, req *DeletePodMonitorReq) (*string, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	name := req.Name
	flag := validateName(name)
	if !flag {
		return nil, fmt.Errorf("校验name参数: %s 是否符合k8s资源名称格式并且长度不大于63位字符不通过", name)
	}
	namespace := req.Namespace
	client, err := servicemonitor.GetMonitoringV1Client(rctx)
	if err != nil {
		return nil, err
	}
	err = client.PodMonitors(namespace).Delete(c, name, metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}
	resp := fmt.Sprintf("删除 Metrics: %s 成功", name)
	return &resp, nil
}

// BatchDeletePodMonitor 批量删除PodMonitor
// @Summary 批量删除PodMonitor
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /pod_monitors/batchdelete [post]
func BatchDeletePodMonitor(c context.Context, req *BatchDeletePodMonitorReq) (*string, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	// 共享集群禁止批量删除
	if rctx.SharedCluster {
		return nil, fmt.Errorf("denied")
	}

	client, err := servicemonitor.GetMonitoringV1Client(rctx)
	if err != nil {
		return nil, err
	}
	for _, v := range req.PodMonitors {
		err = client.PodMonitors(v.Namespace).Delete(c, v.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, err
		}
	}

	resp := fmt.Sprintf("批量删除 Metrics: %s 成功", req)
	return &resp, nil
}

// GetPodMonitor 获取单个PodMonitor
// @Summary 删除PodMonitor
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/:namespace/pod_monitors/:name [get]
func GetPodMonitor(c context.Context, req *GetPodMonitorReq) (*v1.PodMonitor, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	name := req.Name
	flag := validateName(name)
	if !flag {
		return nil, fmt.Errorf("校验name参数: %s 是否符合k8s资源名称格式并且长度不大于63位字符不通过", name)
	}
	namespace := req.Namespace
	client, err := servicemonitor.GetMonitoringV1Client(rctx)
	if err != nil {
		return nil, err
	}
	result, err := client.PodMonitors(namespace).Get(c, name, metav1.GetOptions{})
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
func UpdatePodMonitor(c context.Context, req *UpdatePodMonitorReq) (*any, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	req.Name = req.PathName
	req.Namespace = req.PathNamespace

	flag := req.Validate()

	if !flag {
		return nil, errors.New("参数校验失败")
	}

	client, err := servicemonitor.GetMonitoringV1Client(rctx)
	if err != nil {
		return nil, err
	}

	exist, err := client.PodMonitors(req.Namespace).
		Get(c, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	podMonitor := exist.DeepCopy()
	labels := exist.Labels
	labels["release"] = "po"
	labels["io.tencent.paas.source_type"] = "bcs"
	labels["io.tencent.bcs.service_name"] = req.ServiceName

	params := make(map[string][]string, 0)
	for k, v := range req.Params {
		params[k] = []string{v}
	}
	podMonitor.Labels = labels
	endpoints := make([]v1.PodMetricsEndpoint, 0)
	initEndpoint := v1.PodMetricsEndpoint{
		Port:     req.Port,
		Path:     req.Path,
		Interval: req.Interval,
		Params:   params,
	}

	endpoints = append(endpoints, initEndpoint)
	podMonitor.Spec = v1.PodMonitorSpec{
		PodMetricsEndpoints: endpoints,
		SampleLimit:         uint64(req.SampleLimit),
		Selector: metav1.LabelSelector{
			MatchLabels: req.Selector,
		},
	}
	podMonitorClient := client.PodMonitors(req.Namespace)
	_, err = podMonitorClient.Update(c, podMonitor, metav1.UpdateOptions{})
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
	return len(selector) != 0
}

// validateSampleLimit 校验参数是否合法
func validateSampleLimit(samplelimit int) bool {
	if SampleLimitMax >= samplelimit && samplelimit >= SampleLimitMin {
		return true
	}
	return false
}
