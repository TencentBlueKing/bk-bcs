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

package apiclient

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
)

// PortBindingItemMonitorHelper helper func for monitor client
type PortBindingItemMonitorHelper struct {
	apiCli       IMonitorApiClient
	bcsClusterID string

	IndependentDataID bool
}

// NewPortBindingItemMonitorHelper return new monitor helper
func NewPortBindingItemMonitorHelper() *PortBindingItemMonitorHelper {
	return &PortBindingItemMonitorHelper{
		apiCli:       NewBkmApiClient(),
		bcsClusterID: os.Getenv(constant.EnvNameBkBCSClusterID),

		IndependentDataID: true,
	}
}

// EnsureUptimeCheck ensure uptime check, return true if need retry
func (m *PortBindingItemMonitorHelper) EnsureUptimeCheck(ctx context.Context,
	portBinding *networkextensionv1.PortBinding) bool {
	if portBinding == nil {
		return false
	}
	needRetry := false

	for _, item := range portBinding.Spec.PortBindingList {
		var itemStatus *networkextensionv1.PortBindingStatusItem
		for _, status := range portBinding.Status.PortBindingStatusList {
			if status.GetFullKey() == item.GetFullKey() {
				itemStatus = status
				break
			}
		}

		// 用户关闭了拨测配置
		if !item.UptimeCheck.IsEnabled() {
			if err := m.deleteItemUptimeCheckTask(ctx, itemStatus); err != nil {
				blog.Warnf("delete portBinding[%s/%s] item[%s] uptime check failed, err: %v",
					portBinding.GetNamespace(), portBinding.GetName(), item.GetFullKey(), err.Error())
				needRetry = true
			}
			continue
		}

		// 只由portbindingcontroller.go 中的主逻辑生成status。
		// 仅当portbinding ready (说明clb上的绑定已经完成)时， 才创建对应拨测任务
		// 当portbinding删除 或处于 cleaned / cleaning 状态时， 说明已经开始解绑clb， 提前删除拨测任务
		if itemStatus == nil || itemStatus.Status != constant.PortBindingItemStatusReady {
			blog.Infof("wait portBinding[%s/%s] item[%s] status to be ready...", portBinding.GetNamespace(),
				portBinding.GetName(), item.GetFullKey())
			continue
		}

		taskID, err := m.ensureItemUptimeCheck(ctx, item, itemStatus, portBinding)
		if err != nil {
			blog.Warnf("ensure portBinding[%s/%s] item[%s] uptime check failed, err: %v",
				portBinding.GetNamespace(), portBinding.GetName(), item.GetFullKey(), err.Error())
			var existedID int64 = 0
			if itemStatus.UptimeCheckStatus != nil {
				existedID = itemStatus.UptimeCheckStatus.ID
			}
			itemStatus.UptimeCheckStatus = &networkextensionv1.UptimeCheckTaskStatus{
				Status: networkextensionv1.ListenerStatusNotSynced,
				Msg:    err.Error(),
				ID:     existedID,
			}
			needRetry = true
			continue
		}

		itemStatus.UptimeCheckStatus = &networkextensionv1.UptimeCheckTaskStatus{
			ID:     taskID,
			Status: networkextensionv1.ListenerStatusSynced,
			Msg:    "",
		}
	}

	return needRetry
}

func (m *PortBindingItemMonitorHelper) ensureItemUptimeCheck(ctx context.Context, item *networkextensionv1.PortBindingItem,
	itemStatus *networkextensionv1.PortBindingStatusItem, portBinding *networkextensionv1.PortBinding) (int64, error) {
	cloudTask, err := m.getCloudTask(ctx, item, itemStatus)
	if err != nil {
		return 0, err
	}

	// 有对应拨测任务的情况下，进行对比
	if cloudTask != nil {
		createTaskReq, err1 := m.transUptimeCheckTask(ctx, item, portBinding)
		if err1 != nil {
			return 0, err1
		}
		if m.compareUptimeCheckTask(cloudTask, createTaskReq) {
			if err2 := m.apiCli.DeleteUptimeCheckTask(ctx, cloudTask.ID); err2 != nil {
				return 0, fmt.Errorf("delete uptime check task failed, err: %v", err2)
			}
			resp, err2 := m.apiCli.CreateUptimeCheckTask(ctx, createTaskReq)
			if err2 != nil {
				return 0, fmt.Errorf("create uptime check task failed, err: %v", err2)
			}
			if err2 = m.apiCli.DeployUptimeCheckTask(ctx, resp.Data.ID); err1 != nil {
				return 0, fmt.Errorf("deploy uptime check task failed, err: %v", err2)
			}
			return resp.Data.ID, nil
		}

		return cloudTask.ID, nil
	}

	// 没有对应拨测任务， 则创建
	createTaskReq, err1 := m.transUptimeCheckTask(ctx, item, portBinding)
	if err1 != nil {
		return 0, err1
	}
	resp, err1 := m.apiCli.CreateUptimeCheckTask(ctx, createTaskReq)
	if err1 != nil {
		return 0, fmt.Errorf("create uptime check task failed, err: %v", err1)
	}
	if err = m.apiCli.DeployUptimeCheckTask(ctx, resp.Data.ID); err != nil {
		return 0, fmt.Errorf("deploy uptime check task failed, err: %v", err)
	}
	return resp.Data.ID, nil
}

// DeleteUptimeCheckTask delete uptime check task
func (m *PortBindingItemMonitorHelper) DeleteUptimeCheckTask(ctx context.Context,
	portBinding *networkextensionv1.PortBinding) error {
	if portBinding == nil {
		return nil
	}
	var e error

	for _, itemStatus := range portBinding.Status.PortBindingStatusList {
		if itemStatus.UptimeCheckStatus == nil || itemStatus.UptimeCheckStatus.ID == 0 {
			continue
		}

		if err := m.apiCli.DeleteUptimeCheckTask(ctx, itemStatus.UptimeCheckStatus.ID); err != nil {
			e = fmt.Errorf("delete portBinding[%s/%s] item[%s] uptime check task failed, err: %s",
				portBinding.GetNamespace(), portBinding.GetName(), itemStatus.GetFullKey(), err.Error())
			continue
		}

		itemStatus.UptimeCheckStatus = &networkextensionv1.UptimeCheckTaskStatus{
			ID:     0,
			Status: networkextensionv1.ListenerStatusSynced,
			Msg:    "",
		}
	}

	return e
}

// if portbinding has no uptime check task, skipped and return nil err
func (m *PortBindingItemMonitorHelper) deleteItemUptimeCheckTask(ctx context.Context,
	itemStatus *networkextensionv1.PortBindingStatusItem) error {
	if itemStatus == nil || itemStatus.UptimeCheckStatus == nil || itemStatus.UptimeCheckStatus.ID == 0 {
		return nil
	}

	if err := m.apiCli.DeleteUptimeCheckTask(ctx, itemStatus.UptimeCheckStatus.ID); err != nil {
		blog.Errorf("delete item[%s] uptime check task failed, err: %s",
			itemStatus.GetFullKey(), err.Error())
		return err
	}

	itemStatus.UptimeCheckStatus = &networkextensionv1.UptimeCheckTaskStatus{
		ID:     0,
		Status: networkextensionv1.ListenerStatusSynced,
		Msg:    "",
	}

	return nil
}

// transUptimeCheckTask trans portbinding item to create or update uptime check request
func (m *PortBindingItemMonitorHelper) transUptimeCheckTask(ctx context.Context,
	item *networkextensionv1.PortBindingItem, portBinding *networkextensionv1.PortBinding) (
	*CreateOrUpdateUptimeCheckTaskRequest, error) {
	if portBinding == nil || item == nil || item.UptimeCheck == nil || item.UptimeCheck.IsEnabled() == false {
		return nil, nil
	}

	uptimeCheckConfig := item.UptimeCheck

	nodeIDList, err := m.getNodeIDList(ctx, uptimeCheckConfig.Nodes)
	if err != nil {
		return nil, fmt.Errorf("get node id list for item '%s' failed, err: %v", item.GetFullKey(), err)
	}
	req := &CreateOrUpdateUptimeCheckTaskRequest{
		Protocol:   uptimeCheckConfig.Protocol,
		NodeIDList: nodeIDList,
		Config: Config{
			// IPList:            uptimeCheckConfig.Target, // set target from lb.IP or dnsName
			Period:            uptimeCheckConfig.Period,
			ResponseFormat:    uptimeCheckConfig.ResponseFormat,
			Response:          uptimeCheckConfig.Response,
			Timeout:           uptimeCheckConfig.Timeout,
			Port:              strconv.FormatInt(uptimeCheckConfig.Port, 10), // 对于端口段监听器， 也只拨测首端口
			Request:           uptimeCheckConfig.Request,
			RequestFormat:     uptimeCheckConfig.RequestFormat,
			WaitEmptyResponse: true,
		},
		Name:        m.genUptimeCheckTaskName(item, m.bcsClusterID),
		GroupIDList: []int64{},
	}
	if m.IndependentDataID {
		req.IndependentDataID = true
		req.Labels = map[string]string{
			"bcs_cluster_id":       m.bcsClusterID,
			"bcs_owner_kind":       "PortPool",
			"bcs_owner_name":       item.PoolName,
			"bcs_owner_namespace":  item.PoolNamespace,
			"bcs_source_namespace": portBinding.GetNamespace(),
			"bcs_source_name":      portBinding.GetName(),
		}
	}

	if len(uptimeCheckConfig.URLList) != 0 {
		req.Config.URLList = uptimeCheckConfig.URLList
		req.Config.DNSCheckMode = "all"
		req.Config.TargetIPType = 4
	} else {
		// doto 适配域名化lb
		ipList := make([]string, 0)
		for _, lb := range item.PoolItemLoadBalancers {
			ipList = append(ipList, lb.IPs...)
		}
		req.Config.IPList = ipList
	}

	m.fillCreateOrUpdateReqDefault(req, item)

	return req, nil
}

func (m *PortBindingItemMonitorHelper) fillCreateOrUpdateReqDefault(req *CreateOrUpdateUptimeCheckTaskRequest,
	item *networkextensionv1.PortBindingItem) {
	uptimeCheckConfig := item.UptimeCheck
	// set default
	if req.Config.Period == 0 {
		req.Config.Period = 60
	}
	if req.Config.Timeout == 0 {
		req.Config.Timeout = 3000
	}
	if req.Config.RequestFormat == "" {
		req.Config.RequestFormat = "raw"
	}
	if req.Config.ResponseFormat == "" {
		req.Config.ResponseFormat = "raw"
	}
	if req.Protocol == "" {
		req.Protocol = item.Protocol
	}
	if req.Config.Port == "" || req.Config.Port == "0" {
		req.Config.Port = strconv.Itoa(item.StartPort)
		if uptimeCheckConfig.GetPortDefine() == networkextensionv1.PortDefineLast && item.EndPort != 0 {
			req.Config.Port = strconv.Itoa(item.EndPort)
		}
	}

	req.Config.Method = uptimeCheckConfig.Method
	if req.Config.Method == "" {
		req.Config.Method = httpMethodGet
	}

	// set by uptime check config
	if uptimeCheckConfig.Authorize != nil {
		req.Config.Authorize = &Authorize{
			AuthType: uptimeCheckConfig.Authorize.AuthType,
			AuthConfig: &AuthConfig{
				Token:    uptimeCheckConfig.Authorize.AuthConfig.Token,
				UserName: uptimeCheckConfig.Authorize.AuthConfig.UserName,
				PassWord: uptimeCheckConfig.Authorize.AuthConfig.PassWord,
			},
			InsecureSkipVerify: uptimeCheckConfig.Authorize.InsecureSkipVerify,
		}
	}
	if uptimeCheckConfig.Body != nil {
		req.Config.Body = &Body{
			DataType:    uptimeCheckConfig.Body.DataType,
			Content:     uptimeCheckConfig.Body.Content,
			ContentType: uptimeCheckConfig.Body.ContentType,
		}

		for _, param := range uptimeCheckConfig.Body.Params {
			req.Config.Body.Params = append(req.Config.Body.Params, &Params{
				IsEnabled: param.IsEnabled,
				Key:       param.Key,
				Value:     param.Value,
				Desc:      param.Desc,
			})
		}
	}

	if len(uptimeCheckConfig.QueryParams) != 0 {
		for _, param := range uptimeCheckConfig.QueryParams {
			req.Config.QueryParams = append(req.Config.QueryParams, &Params{
				IsEnabled: param.IsEnabled,
				Key:       param.Key,
				Value:     param.Value,
				Desc:      param.Desc,
			})
		}
	}

	if len(uptimeCheckConfig.Headers) != 0 {
		for _, param := range uptimeCheckConfig.Headers {
			req.Config.Headers = append(req.Config.Headers, &Params{
				IsEnabled: param.IsEnabled,
				Key:       param.Key,
				Value:     param.Value,
				Desc:      param.Desc,
			})
		}
	}

	if uptimeCheckConfig.ResponseCode != "" {
		req.Config.ResponseCode = uptimeCheckConfig.ResponseCode
	}
}

// getNodeIDList nodes can be node ip or node name
func (m *PortBindingItemMonitorHelper) getNodeIDList(ctx context.Context, nodes []string) ([]int64, error) {
	listNodeResp, err := m.apiCli.ListNode(ctx)
	if err != nil {
		return nil, fmt.Errorf("list check node failed, err :%v", err)
	}
	nodeNameMap := make(map[string]int64)
	nodeIPMap := make(map[string]int64)
	for _, node := range listNodeResp.Data {
		nodeNameMap[node.Name] = node.ID
		nodeIPMap[node.IP] = node.ID
	}

	res := make([]int64, 0, len(nodes))
	for _, node := range nodes {
		if ip, ok := nodeIPMap[node]; ok {
			res = append(res, ip)
			continue
		}

		if ip, ok := nodeNameMap[node]; ok {
			res = append(res, ip)
			continue
		}
		return nil, fmt.Errorf("not found node %s", node)
	}

	return res, nil
}

// return true if need update
func (m *PortBindingItemMonitorHelper) compareUptimeCheckTask(cloudTask *UptimeCheckTask, createReq *CreateOrUpdateUptimeCheckTaskRequest) bool {
	if cloudTask.Name != createReq.Name {
		return true
	}
	if len(cloudTask.Nodes) != len(createReq.NodeIDList) {
		return true
	}

	for _, node := range cloudTask.Nodes {
		found := false
		for _, nodeID := range createReq.NodeIDList {
			if node.ID == nodeID {
				found = true
				break
			}
		}

		if !found {
			return true
		}
	}

	if !compareStringList(cloudTask.Config.IPList, createReq.Config.IPList) {
		return true
	}

	if m.IndependentDataID {
		if cloudTask.IndependentDataID != createReq.IndependentDataID || !reflect.DeepEqual(cloudTask.Labels,
			createReq.Labels) {
			return true
		}
	}

	return false
}

func (m *PortBindingItemMonitorHelper) genUptimeCheckTaskName(item *networkextensionv1.PortBindingItem,
	bcsClusterID string) string {
	port := item.StartPort
	if item.UptimeCheck.GetPortDefine() == networkextensionv1.PortDefineLast && item.EndPort != 0 {
		port = item.EndPort
	}
	return fmt.Sprintf("bcs-%s/%s/%s/%d/%s", item.PoolNamespace, item.PoolName, item.PoolItemName, port, bcsClusterID)
}

func (m *PortBindingItemMonitorHelper) getCloudTask(ctx context.Context, item *networkextensionv1.PortBindingItem,
	status *networkextensionv1.PortBindingStatusItem) (*UptimeCheckTask, error) {
	var cloudTask *UptimeCheckTask
	taskName := m.genUptimeCheckTaskName(item, m.bcsClusterID)
	// 优先使用状态中记录的ID寻找， 如果没有记录的话， 使用Name再次确认
	if status.UptimeCheckStatus != nil && status.UptimeCheckStatus.ID != 0 {
		taskID := status.UptimeCheckStatus.ID
		taskResp, err := m.apiCli.ListUptimeCheckTask(ctx, &ListUptimeCheckRequest{Id: taskID})
		if err != nil {
			return nil, err
		}
		if len(taskResp.Data) > 1 {
			return nil, fmt.Errorf("get uptime check task by ID[%d] more than 1, task count: %d",
				taskID, len(taskResp.Data))
		}
		if len(taskResp.Data) == 1 {
			cloudTask = taskResp.Data[0]
		}
	}
	if cloudTask == nil {
		taskResp, err := m.apiCli.ListUptimeCheckTask(ctx, &ListUptimeCheckRequest{Name: taskName})
		if err != nil {
			return nil, err
		}
		if len(taskResp.Data) > 1 {
			return nil, fmt.Errorf("get uptime check task by Name[%s] more than 1, task count: %d",
				taskName, len(taskResp.Data))
		}
		if len(taskResp.Data) == 1 {
			cloudTask = taskResp.Data[0]
		}
	}

	return cloudTask, nil
}

// compareStringList return true if given lists have same elements
func compareStringList(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for _, item := range a {
		found := false
		for _, target := range b {
			if item == target {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
