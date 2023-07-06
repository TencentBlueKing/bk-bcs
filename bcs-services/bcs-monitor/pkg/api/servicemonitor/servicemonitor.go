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

package service_monitor

import (
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

func GetMonitoringV1Client(c *rest.Context) (monitoringv1.MonitoringV1Interface, error) {
	clusterId := c.Param("clusterId")
	bcsConf := k8sclient.GetBCSConf()
	host := fmt.Sprintf("%s/clusters/%s", bcsConf.Host, clusterId)
	k8sconfig := &k8srest.Config{
		Host:        host,
		BearerToken: bcsConf.Token,
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
func ListServiceMonitors(c *rest.Context) (interface{}, error) {

	limit := c.Query("limit")
	offset := c.Query("offset")
	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = c.Param("namespace")
	}
	client, err := GetMonitoringV1Client(c)
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
	data, err := client.ServiceMonitors(namespace).List(c.Context, listOps)
	if err != nil {
		return nil, err
	}
	serviceMonitors := make([]*v1.ServiceMonitor, 0)
	for _, v := range data.Items {
		serviceMonitors = append(serviceMonitors, v)
	}
	return serviceMonitors, nil
}

// CreateServiceMonitor 创建ServiceMonitor
// @Summary ServiceMonitor列表数据
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /service_monitors/create/namespaces/:namespace/servicemonitors/:name [post]
func CreateServiceMonitor(c *rest.Context) (interface{}, error) {
	serviceMonitorReq := &CreateServiceMonitorReq{}
	if err := c.ShouldBindJSON(serviceMonitorReq); err != nil {
		return nil, err
	}
	serviceMonitorReq.Namespace = c.Param("namespace")

	flag := serviceMonitorReq.Validate()

	if !flag {
		return nil, errors.New("参数校验失败")
	}

	client, err := GetMonitoringV1Client(c)
	if err != nil {
		return nil, err
	}
	params := make(map[string][]string, 0)
	for k, v := range serviceMonitorReq.Params {
		params[k] = []string{v}
	}

	serviceMonitor := &v1.ServiceMonitor{}
	labels := map[string]string{
		"release":                     "po",
		"io.tencent.paas.source_type": "bcs",
		"io.tencent.bcs.service_name": serviceMonitorReq.ServiceName,
	}

	serviceMonitor.ObjectMeta = metav1.ObjectMeta{
		Labels:    labels,
		Name:      serviceMonitorReq.Name,
		Namespace: serviceMonitorReq.Namespace,
	}
	endpoints := make([]v1.Endpoint, 0)
	initEndpoint := v1.Endpoint{
		Port:     serviceMonitorReq.Port,
		Path:     serviceMonitorReq.Path,
		Interval: serviceMonitorReq.Interval,
		Params:   params,
	}

	endpoints = append(endpoints, initEndpoint)
	serviceMonitor.Spec = v1.ServiceMonitorSpec{
		Endpoints:   endpoints,
		SampleLimit: uint64(serviceMonitorReq.SampleLimit),
		Selector: metav1.LabelSelector{
			MatchLabels: serviceMonitorReq.Selector,
		},
	}

	_, err = client.ServiceMonitors(serviceMonitorReq.Namespace).Create(c.Context, serviceMonitor, metav1.CreateOptions{})

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
func DeleteServiceMonitor(c *rest.Context) (interface{}, error) {
	name := c.Param("name")
	flag := validateName(name)
	if !flag {
		return nil, fmt.Errorf("校验name参数: %s 是否符合k8s资源名称格式并且长度不大于63位字符不通过", name)
	}
	namespace := c.Param("namespace")
	client, err := GetMonitoringV1Client(c)
	if err != nil {
		return nil, err
	}
	err = client.ServiceMonitors(namespace).Delete(c, name, metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("删除 Metrics: %s 成功", name), nil
}

// BatchDeleteServiceMonitor 批量删除ServiceMonitor
// @Summary 批量删除ServiceMonitor
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /service_monitors/batchdelete [delete]
func BatchDeleteServiceMonitor(c *rest.Context) (interface{}, error) {
	serviceMonitorDelReq := &BatchDeleteServiceMonitorReq{}
	if err := c.ShouldBindJSON(serviceMonitorDelReq); err != nil {
		return nil, err
	}
	client, err := GetMonitoringV1Client(c)
	if err != nil {
		return nil, err
	}
	for _, v := range serviceMonitorDelReq.ServiceMonitors {
		err = client.ServiceMonitors(v.Namespace).Delete(c, v.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, err
		}
	}

	return fmt.Sprintf("批量删除 Metrics: %s 成功", serviceMonitorDelReq), nil
}

// GetServiceMonitor 获取单个ServiceMonitor
// @Summary 删除ServiceMonitor
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /service_monitors/namespaces/:namespace/servicemonitors/:name [delete]
func GetServiceMonitor(c *rest.Context) (interface{}, error) {
	name := c.Param("name")
	flag := validateName(name)
	if !flag {
		return nil, fmt.Errorf("校验name参数: %s 是否符合k8s资源名称格式并且长度不大于63位字符不通过", name)
	}
	namespace := c.Param("namespace")
	client, err := GetMonitoringV1Client(c)
	if err != nil {
		return nil, err
	}
	result, err := client.ServiceMonitors(namespace).Get(c.Context, name, metav1.GetOptions{})
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
func UpdateServiceMonitor(c *rest.Context) (interface{}, error) {
	serviceMonitorReq := &UpdateServiceMonitorReq{}
	if err := c.ShouldBindJSON(serviceMonitorReq); err != nil {
		return nil, err
	}
	serviceMonitorReq.Name = c.Param("name")
	serviceMonitorReq.Namespace = c.Param("namespace")

	flag := serviceMonitorReq.Validate()

	if !flag {
		return nil, errors.New("参数校验失败")
	}

	client, err := GetMonitoringV1Client(c)
	if err != nil {
		return nil, err
	}

	exist, err := client.ServiceMonitors(serviceMonitorReq.Namespace).
		Get(c.Context, serviceMonitorReq.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	serviceMonitor := exist.DeepCopy()
	labels := exist.Labels
	labels["release"] = "po"
	labels["io.tencent.paas.source_type"] = "bcs"
	labels["io.tencent.bcs.service_name"] = serviceMonitorReq.ServiceName

	params := make(map[string][]string, 0)
	for k, v := range serviceMonitorReq.Params {
		params[k] = []string{v}
	}
	serviceMonitor.Labels = labels
	endpoints := make([]v1.Endpoint, 0)
	initEndpoint := v1.Endpoint{
		Port:     serviceMonitorReq.Port,
		Path:     serviceMonitorReq.Path,
		Interval: serviceMonitorReq.Interval,
		Params:   params,
	}

	endpoints = append(endpoints, initEndpoint)
	serviceMonitor.Spec = v1.ServiceMonitorSpec{
		Endpoints:   endpoints,
		SampleLimit: uint64(serviceMonitorReq.SampleLimit),
		Selector: metav1.LabelSelector{
			MatchLabels: serviceMonitorReq.Selector,
		},
	}
	serviceMonitorClient := client.ServiceMonitors(serviceMonitorReq.Namespace)
	_, err = serviceMonitorClient.Update(c.Context, serviceMonitor, metav1.UpdateOptions{})
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
	//第一个正则表达式 ^[a-z][-a-z0-9]*$ 匹配以小写字母开头，后跟任意数量的小写字母、数字和短横线的字符串。这个正则表达式不允许字符串末尾有短横线，因为 [-a-z0-9]* 匹配任意数量的小写字母、数字和短横线，但它不需要匹配任何字符。
	//
	//第二个正则表达式 ^[a-z0-9]([-a-z0-9]*[a-z0-9])?$ 匹配以小写字母或数字开头，后跟任意数量的小写字母、数字和短横线的字符串。它允许字符串末尾有短横线，因为 ([-a-z0-9]*[a-z0-9])? 匹配任意数量的小写字母、数字和短横线，后跟一个小写字母或数字，这个组合出现零次或一次，因此字符串末尾可以是短横线或小写字母或数字。
	//              ^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	//因此，虽然这两个正则表达式都匹配以小写字母或数字开头，后跟任意数量的小写字母、数字和短横线的字符串，但是第二个正则表达式允许字符串末尾有短横线，因此与第一个正则表达式不完全相同。
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
	if SM_SAMPLE_LIMIT_MAX >= samplelimit && samplelimit >= SM_SAMPLE_LIMIT_MIN {
		return true
	}
	return false
}
