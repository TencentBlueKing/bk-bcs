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
	"strconv"
	"strings"

	gocache "github.com/patrickmn/go-cache"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// MonitorHelper helper func for monitor client
type MonitorHelper struct {
	apiCli       IMonitorApiClient
	lbIDCache    *gocache.Cache
	bcsClusterID string
}

// NewMonitorHelper return new monitor helper
func NewMonitorHelper(lbIDCache *gocache.Cache) *MonitorHelper {
	return &MonitorHelper{
		apiCli:       NewBkmApiClient(),
		lbIDCache:    lbIDCache,
		bcsClusterID: os.Getenv(constant.EnvNameBkBCSClusterID),
	}
}

// EnsureUptimeCheck ensure uptime check
func (m *MonitorHelper) EnsureUptimeCheck(ctx context.Context, listener *networkextensionv1.Listener) (int64, error) {
	if listener == nil {
		return 0, nil
	}

	if !listener.IsUptimeCheckEnable() {
		return 0, nil
	}
	taskName := genUptimeCheckTaskName(listener, m.bcsClusterID)
	var cloudTask *UptimeCheckTask

	taskResp, err := m.apiCli.ListUptimeCheckTask(ctx)
	if err != nil {
		return 0, err
	}
	for _, task := range taskResp.Data {
		if listener.GetUptimeCheckTaskStatus().ID != 0 && task.ID == listener.GetUptimeCheckTaskStatus().ID {
			cloudTask = task
			break
		}
		if task.Name == taskName {
			cloudTask = task
			break
		}
	}
	// 1. 有task， 但目前targetGroup为空 -> 需要删除对应拨测任务
	// 2. 有task， targetGroup有值 -> 对应更新拨测任务（判断是否要更新）
	// 3. 没task， targetGroup没值 -> 无事发生
	// 4. 没task， targetGroup有值 -> 创建拨测任务
	if cloudTask != nil {
		if listener.IsEmptyTargetGroup() {
			blog.Info("listener '%s/%s' empty targetGroup, delete uptime check task[%d]... ", listener.GetNamespace(),
				listener.GetName(), cloudTask.ID)
			if err = m.apiCli.DeleteUptimeCheckTask(ctx, cloudTask.ID); err != nil {
				return 0, fmt.Errorf("delete uptime check task'%d' faield ,err: %v", cloudTask.ID, err)
			}
			return 0, nil
		}

		createTaskReq, err1 := m.transUptimeCheckTask(ctx, listener)
		if err1 != nil {
			return 0, err1
		}
		if m.compareUptimeCheckTask(cloudTask, createTaskReq) {
			resp, err2 := m.apiCli.UpdateUptimeCheckTask(ctx, createTaskReq)
			if err2 != nil {
				return 0, fmt.Errorf("update uptime check task failed, err: %v", err2)
			}
			if err2 = m.apiCli.DeployUptimeCheckTask(ctx, resp.Data.ID); err1 != nil {
				return 0, fmt.Errorf("deploy uptime check task failed, err: %v", err2)
			}
		}

		return cloudTask.ID, nil
	}

	if listener.IsEmptyTargetGroup() {
		blog.Info("listener '%s/%s' empty targetGroup, skip uptime check task... ", listener.GetNamespace(),
			listener.GetName())
		return 0, nil
	}

	createTaskReq, err := m.transUptimeCheckTask(ctx, listener)
	if err != nil {
		return 0, err
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
func (m *MonitorHelper) DeleteUptimeCheckTask(ctx context.Context, listener *networkextensionv1.Listener) error {
	if listener.GetUptimeCheckTaskStatus().ID == 0 {
		return nil
	}
	return m.apiCli.DeleteUptimeCheckTask(ctx, listener.GetUptimeCheckTaskStatus().ID)
}

func (m *MonitorHelper) transUptimeCheckTask(ctx context.Context, listener *networkextensionv1.Listener) (
	*CreateOrUpdateUptimeCheckTaskRequest, error) {
	if !listener.IsUptimeCheckEnable() {
		return nil, nil
	}

	uptimeCheckConfig := listener.Spec.ListenerAttribute.UptimeCheck

	nodeIDList, err := m.getNodeIDList(ctx, uptimeCheckConfig.Nodes)
	if err != nil {
		return nil, fmt.Errorf("get node id list for listener '%s/%s' failed, err: %v", listener.GetNamespace(),
			listener.GetName(), err)
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
		Name:        genUptimeCheckTaskName(listener, m.bcsClusterID),
		GroupIDList: []int64{},
	}

	if len(uptimeCheckConfig.URLList) != 0 {
		req.Config.URLList = uptimeCheckConfig.URLList
		req.Config.DNSCheckMode = "all"
		req.Config.TargetIPType = 4
	} else {
		// get target info from listener's lbID
		var lbObj *cloud.LoadBalanceObject
		obj, ok := m.lbIDCache.Get(listener.GetRegion() + ":" + listener.Spec.LoadbalancerID)
		if ok {
			if lbObj, ok = obj.(*cloud.LoadBalanceObject); !ok {
				return nil, fmt.Errorf("get obj from lb id cache is not LoadBalanceObject")
			}
		} else {
			// LB信息通常会被缓存。 当Controller重启时，可能出现缓存更新不及时，导致找不到LB信息
			return nil, fmt.Errorf("wait for listener's lb[%s/%s] be cached", listener.GetRegion(), listener.Spec.LoadbalancerID)
		}

		// DNS / rules.domain
		if lbObj.DNSName != "" {
			if listener.Spec.Protocol == constant.ProtocolHTTPS {
				req.Config.URLList = []string{fmt.Sprintf("https://%s", lbObj.DNSName)}
			} else {
				req.Config.URLList = []string{fmt.Sprintf("http://%s", lbObj.DNSName)}
			}
			req.Config.DNSCheckMode = "all"
			req.Config.TargetIPType = 4
		} else if len(lbObj.IPs) != 0 {
			req.Config.IPList = lbObj.IPs
		}
	}

	m.fillCreateOrUpdateReqDefault(req, listener)

	return req, nil
}

func (m *MonitorHelper) fillCreateOrUpdateReqDefault(req *CreateOrUpdateUptimeCheckTaskRequest,
	listener *networkextensionv1.Listener) {
	uptimeCheckConfig := listener.Spec.ListenerAttribute.UptimeCheck
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
		if strings.ToUpper(listener.Spec.Protocol) == constant.ProtocolHTTPS {
			req.Protocol = constant.ProtocolHTTP
		} else {
			req.Protocol = listener.Spec.Protocol
		}
	}
	if req.Config.Port == "" || req.Config.Port == "0" {
		req.Config.Port = strconv.Itoa(listener.Spec.Port)
	}

	req.Config.Method = uptimeCheckConfig.Method
	if req.Config.Method == "" {
		req.Config.Method = httpMethodGet
	}

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
func (m *MonitorHelper) getNodeIDList(ctx context.Context, nodes []string) ([]int64, error) {
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
func (m *MonitorHelper) compareUptimeCheckTask(cloudTask *UptimeCheckTask, createReq *CreateOrUpdateUptimeCheckTaskRequest) bool {
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

	// todo 增加协议/端口的匹配判断
	return false
}

func genUptimeCheckTaskName(listener *networkextensionv1.Listener, bcsClusterID string) string {
	return fmt.Sprintf("bcs-%s-%s-%d/%s/%s", listener.Spec.LoadbalancerID, listener.Spec.Protocol,
		listener.Spec.Port, bcsClusterID, listener.GetListenerSourceNamespace())
}
