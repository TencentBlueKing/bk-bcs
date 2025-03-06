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

// Package servicemonitor prometheus service monitor
package servicemonitor

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1client "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8srest "k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

// GetMonitoringV1Client get monitoring client
func GetMonitoringV1Client(c *rest.Context) (monitoringv1.MonitoringV1Interface, error) {
	clusterId := c.ClusterId
	bcsConf := k8sclient.GetBCSConf()
	host := fmt.Sprintf("%s/clusters/%s", bcsConf.Host, clusterId)
	k8sconfig := &k8srest.Config{
		Host:            host,
		BearerToken:     bcsConf.Token,
		TLSClientConfig: k8srest.TLSClientConfig{Insecure: true},
	}
	config, err := v1client.NewForConfig(k8sconfig)
	if err != nil {
		return nil, err
	}
	return config.MonitoringV1(), nil
}

// ListServiceMonitors 获取ServiceMonitor列表数据
// @Summary ServiceMonitor列表数据
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /service_monitors/namespaces/:namespace [get]
func ListServiceMonitors(c context.Context, req *ListServiceMonitorsReq) (*[]*v1.ServiceMonitor, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	serviceMonitors := make([]*v1.ServiceMonitor, 0)
	// 共享集群不展示列表
	if rctx.SharedCluster {
		return &serviceMonitors, nil
	}
	limit := req.Limit
	offset := req.Offset
	namespace := req.Namespace
	if namespace == "" {
		namespace = req.PathNamespace
	}
	client, err := GetMonitoringV1Client(rctx)
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
	data, err := client.ServiceMonitors(namespace).List(c, listOps)
	if err != nil {
		return nil, err
	}
	serviceMonitors = append(serviceMonitors, data.Items...)
	return &serviceMonitors, nil
}

// CreateServiceMonitor 创建ServiceMonitor
// @Summary ServiceMonitor列表数据
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /service_monitors/create/namespaces/:namespace/servicemonitors/:name [post]
func CreateServiceMonitor(c context.Context, req *CreateServiceMonitorReq) (*v1.ServiceMonitor, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	req.Namespace = req.PathNamespace

	flag := req.Validate()

	if !flag {
		return nil, errors.New("参数校验失败")
	}

	client, err := GetMonitoringV1Client(rctx)
	if err != nil {
		return nil, err
	}
	params := make(map[string][]string, 0)
	for k, v := range req.Params {
		params[k] = []string{v}
	}

	serviceMonitor := &v1.ServiceMonitor{}
	labels := map[string]string{
		"release":                     "po",
		"io.tencent.paas.source_type": "bcs",
		"io.tencent.bcs.service_name": req.ServiceName,
	}

	serviceMonitor.ObjectMeta = metav1.ObjectMeta{
		Labels:    labels,
		Name:      req.Name,
		Namespace: req.Namespace,
	}
	endpoints := make([]v1.Endpoint, 0)
	initEndpoint := v1.Endpoint{
		Port:     req.Port,
		Path:     req.Path,
		Interval: req.Interval,
		Params:   params,
	}

	endpoints = append(endpoints, initEndpoint)
	serviceMonitor.Spec = v1.ServiceMonitorSpec{
		Endpoints:   endpoints,
		SampleLimit: uint64(req.SampleLimit),
		Selector: metav1.LabelSelector{
			MatchLabels: req.Selector,
		},
	}

	_, err = client.ServiceMonitors(req.Namespace).Create(c, serviceMonitor, metav1.CreateOptions{})

	if err != nil {
		return nil, err
	}
	return nil, nil
}

// DeleteServiceMonitor 删除ServiceMonitor
// @Summary 删除ServiceMonitor
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /service_monitors/namespaces/:namespace/servicemonitors/:name [delete]
func DeleteServiceMonitor(c context.Context, req *DeleteServiceMonitorReq) (*string, error) {
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
	client, err := GetMonitoringV1Client(rctx)
	if err != nil {
		return nil, err
	}
	err = client.ServiceMonitors(namespace).Delete(c, name, metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}

	resp := fmt.Sprintf("删除 Metrics: %s 成功", name)
	return &resp, nil
}

// BatchDeleteServiceMonitor 批量删除ServiceMonitor
// @Summary 批量删除ServiceMonitor
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /service_monitors/batchdelete [delete]
func BatchDeleteServiceMonitor(
	c context.Context, req *BatchDeleteServiceMonitorReq) (*string, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	// 共享集群禁止批量删除
	if rctx.SharedCluster {
		return nil, fmt.Errorf("denied")
	}

	client, err := GetMonitoringV1Client(rctx)
	if err != nil {
		return nil, err
	}
	for _, v := range req.ServiceMonitors {
		err = client.ServiceMonitors(v.Namespace).Delete(c, v.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, err
		}
	}

	resp := fmt.Sprintf("批量删除 Metrics: %s 成功", req)
	return &resp, nil
}

// GetServiceMonitor 获取单个ServiceMonitor
// @Summary 删除ServiceMonitor
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /service_monitors/namespaces/:namespace/servicemonitors/:name [delete]
func GetServiceMonitor(c context.Context, req *GetServiceMonitorReq) (*v1.ServiceMonitor, error) {
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
	client, err := GetMonitoringV1Client(rctx)
	if err != nil {
		return nil, err
	}
	result, err := client.ServiceMonitors(namespace).Get(c, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateServiceMonitor 创建ServiceMonitor
// @Summary ServiceMonitor列表数据
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /service_monitors/update/namespaces/:namespace/servicemonitors/:name [put]
func UpdateServiceMonitor(c context.Context, req *UpdateServiceMonitorReq) (*v1.ServiceMonitor, error) {
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

	client, err := GetMonitoringV1Client(rctx)
	if err != nil {
		return nil, err
	}

	exist, err := client.ServiceMonitors(req.Namespace).
		Get(c, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	serviceMonitor := exist.DeepCopy()
	labels := exist.Labels
	labels["release"] = "po"
	labels["io.tencent.paas.source_type"] = "bcs"
	labels["io.tencent.bcs.service_name"] = req.ServiceName

	params := make(map[string][]string, 0)
	for k, v := range req.Params {
		params[k] = []string{v}
	}
	serviceMonitor.Labels = labels
	endpoints := make([]v1.Endpoint, 0)
	initEndpoint := v1.Endpoint{
		Port:     req.Port,
		Path:     req.Path,
		Interval: req.Interval,
		Params:   params,
	}

	endpoints = append(endpoints, initEndpoint)
	serviceMonitor.Spec = v1.ServiceMonitorSpec{
		Endpoints:   endpoints,
		SampleLimit: uint64(req.SampleLimit),
		Selector: metav1.LabelSelector{
			MatchLabels: req.Selector,
		},
	}
	serviceMonitorClient := client.ServiceMonitors(req.Namespace)
	_, err = serviceMonitorClient.Update(c, serviceMonitor, metav1.UpdateOptions{})
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
func validateSampleLimit(samplelimit int) bool { // nolint
	if SampleLimitMax >= samplelimit && samplelimit >= SampleLimitMin {
		return true
	}
	return false
}
