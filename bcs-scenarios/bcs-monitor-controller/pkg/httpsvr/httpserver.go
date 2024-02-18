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

package httpsvr

import (
	"context"
	"fmt"
	"time"

	"github.com/emicklei/go-restful"
	"gopkg.in/yaml.v3"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	monitorextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
)

// HttpServerClient http server client
type HttpServerClient struct {
	Mgr manager.Manager
}

// InitRouters init router
func InitRouters(ws *restful.WebService, httpServerClient *HttpServerClient) {
	ws.Route(ws.GET("/api/v1/monitor/{biz_id}").To(httpServerClient.ListAppMonitors))
	ws.Route(ws.POST("/api/v1/monitor/{biz_id}/{scenario}").To(httpServerClient.CreateOrUpdateAppMonitor))
	ws.Route(ws.DELETE("/api/v1/monitor/{biz_id}/{scenario}").To(httpServerClient.DeleteAppMonitor))

	// ws.Route(ws.GET("/api/v1/list_argo_repo").To(httpServerClient.GetArgoRepo))
}

// func (h *HttpServerClient) GetArgoRepo(request *restful.Request, response *restful.Response) {
// 	argoDB, _, err := repo.NewArgoDB(context.Background(), "default")
// 	if err != nil {
// 		blog.Errorf("connect argo failed, err: %s", err.Error())
// 		_, _ = response.Write(CreateResponseData(err, "", nil))
// 		return
// 	}
//
// 	repos, err := argoDB.ListRepositories(context.Background())
// 	if err != nil {
// 		blog.Errorf("list argo repo failed, err: %s", err.Error())
// 		_, _ = response.Write(CreateResponseData(err, "", nil))
// 		return
// 	}
// 	_, _ = response.Write(CreateResponseData(nil, "", utils.ToJsonString(repos)))
// 	return
// }

// ListAppMonitors list app monitor
func (h *HttpServerClient) ListAppMonitors(request *restful.Request, response *restful.Response) {
	// InstalledScenarioInfo install
	type InstalledScenarioInfo struct {
		Name    string `json:"name"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	type Resp struct {
		InstallScenario []InstalledScenarioInfo `json:"install_scenario"`
	}
	bizID := request.PathParameter("biz_id")
	appMonitorList := &monitorextensionv1.AppMonitorList{}
	// LabelSelectorAsSelector converts the LabelSelector api type into a struct that implements
	// labels.Selector
	// Note: This function should be kept in sync with the selector methods in pkg/labels/selector.go
	selector, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(map[string]string{
		monitorextensionv1.LabelKeyForBizID: bizID,
	}))
	if err != nil {
		blog.Errorf("build selector failed, err: %s", err.Error())
		_, _ = response.Write(CreateResponseData(fmt.Errorf("build selector failed, err: %w", err), "", nil))
		return
	}
	err = h.Mgr.GetClient().List(context.Background(), appMonitorList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		blog.Errorf("read api-server failed, err: %s", err.Error())
		_, _ = response.Write(CreateResponseData(fmt.Errorf("read api-server failed, err: %w", err), "", nil))
		return
	}

	infoList := make([]InstalledScenarioInfo, 0)
	for _, appMonitor := range appMonitorList.Items {
		infoList = append(infoList, InstalledScenarioInfo{Name: appMonitor.Spec.Scenario,
			Status:  string(appMonitor.Status.SyncStatus.State),
			Message: appMonitor.Status.SyncStatus.Message})
	}

	_, _ = response.Write(CreateResponseData(nil, "", Resp{InstallScenario: infoList}))
}

// CreateOrUpdateAppMonitor create or update app monitor
func (h *HttpServerClient) CreateOrUpdateAppMonitor(request *restful.Request, response *restful.Response) {
	// Req entityPointer req
	type Req struct {
		BizID    string `json:"biz_id"`
		Scenario string `json:"scenario"`
		Values   string `json:"values"`
	}
	req := &Req{}
	if err := request.ReadEntity(req); err != nil {
		_, _ = response.Write(CreateResponseData(fmt.Errorf("read body params 'values'failed, err: %w", err), "", nil))
		return
	}
	req.BizID = request.PathParameter("biz_id")
	req.Scenario = request.PathParameter("scenario")

	blog.Infof("bizID: %s, sce: %s,values: %s", req.BizID, req.Scenario, req.Values)

	if req.BizID == "" || req.Scenario == "" || req.Values == "" {
		_, _ = response.Write(CreateResponseData(fmt.Errorf("empty param biz_id or scenario or values"), "", nil))
		return
	}

	namespacedName, err := h.doCreateOrUpdateAppMonitor(req.BizID, req.Scenario, req.Values)
	if err != nil {
		blog.Errorf("doCreateOrUpdateAppMonitor failed, bizID[%s], scenario[%s], values[%s], err: %s", req.BizID,
			req.Scenario, req.Values, err.Error())
		_, _ = response.Write(CreateResponseData(fmt.Errorf("doCreateOrUpdateAppMonitor failed, err: %w", err), "",
			nil))
		return
	}

	// for loop? chan?
	var appMonitor monitorextensionv1.AppMonitor
	for {
		if inErr := h.Mgr.GetAPIReader().Get(request.Request.Context(), *namespacedName, &appMonitor); inErr != nil {
			blog.Errorf("get app monitor '%s/%s' failed: %s", namespacedName.Namespace, namespacedName.Name,
				inErr.Error())
			_, _ = response.Write(CreateResponseData(fmt.Errorf("get app monitor '%s/%s' failed: %w",
				namespacedName.Namespace, namespacedName.Name, inErr), "", nil))
		}

		if appMonitor.Status.SyncStatus.State == monitorextensionv1.SyncStateFailed {
			_, _ = response.Write(CreateResponseData(fmt.Errorf("sync failed, %s",
				appMonitor.Status.SyncStatus.Message), "sync failed", struct{}{}))
			return
		}
		if appMonitor.Status.SyncStatus.State == monitorextensionv1.SyncStateCompleted {
			_, _ = response.Write(CreateResponseData(nil, "success", struct{}{}))
			return
		}

		time.Sleep(time.Second)
	}
}

// DeleteAppMonitor delete app monitor
func (h *HttpServerClient) DeleteAppMonitor(request *restful.Request, response *restful.Response) {
	// Req request
	type Req struct {
		BizID    string `json:"biz_id"`
		Scenario string `json:"scenario"`
	}
	req := &Req{}
	req.BizID = request.PathParameter("biz_id")
	req.Scenario = request.PathParameter("scenario")
	blog.Infof("bizID: %s, sce: %s", req.BizID, req.Scenario)

	if req.BizID == "" || req.Scenario == "" {
		_, _ = response.Write(CreateResponseData(fmt.Errorf("empty param biz_id or scenario"), "", nil))
		return
	}

	if err := h.doDeleteAppMonitor(req.BizID, req.Scenario); err != nil {
		blog.Errorf("%s", err.Error())
		_, _ = response.Write(CreateResponseData(err, "", nil))
		return
	}
	_, _ = response.Write(CreateResponseData(nil, "success", struct{}{}))
	return
}

// do Create Or Update App Monitor
func (h *HttpServerClient) doCreateOrUpdateAppMonitor(bizID, scenario, values string) (*k8stypes.NamespacedName, error) {
	var (
		appMonitor    *monitorextensionv1.AppMonitor
		foundPrevious bool
	)
	appMonitorList := &monitorextensionv1.AppMonitorList{}
	selector, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(map[string]string{
		monitorextensionv1.LabelKeyForBizID:        bizID,
		monitorextensionv1.LabelKeyForScenarioName: scenario,
	}))
	if err != nil {
		return nil, fmt.Errorf("build selector failed, err: %w", err)
	}
	err = h.Mgr.GetClient().List(context.Background(), appMonitorList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, fmt.Errorf("read api-server failed, err: %s", err)
	}

	if len(appMonitorList.Items) != 0 {
		if len(appMonitorList.Items) > 1 {
			return nil, fmt.Errorf("unknown error, multi scenario found")
		}

		appMonitor = &appMonitorList.Items[0]
		foundPrevious = true
	} else {
		appMonitor = &monitorextensionv1.AppMonitor{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name:      genAppMonitorName(bizID, scenario),
				Namespace: monitorextensionv1.DefaultNamespace,
			},
		}
		foundPrevious = false
	}
	appMonitor.Status.SyncStatus.State = monitorextensionv1.SyncStateNeedReSync
	appMonitor.Spec.Scenario = scenario
	appMonitor.Spec.BizId = bizID
	appMonitor.Spec.Override = true

	if err = yaml.Unmarshal([]byte(values), &appMonitor.Spec); err != nil {
		return nil, fmt.Errorf("json unmarshal values failed, biz_id[%s], scenario[%s], err: %w", bizID, scenario,
			err)
	}

	if foundPrevious {
		blog.Infof("update previous AppMonitor'%s/%s'", appMonitor.GetNamespace(), appMonitor.GetName())
		if err = h.Mgr.GetClient().Update(context.Background(), appMonitor); err != nil {
			return nil, fmt.Errorf("update appmonitor '%s/%s' failed", appMonitor.GetNamespace(), appMonitor.GetName())
		}
	} else {
		blog.Infof("create AppMonitor'%s/%s'", appMonitor.GetNamespace(), appMonitor.GetName())
		if err = h.Mgr.GetClient().Create(context.Background(), appMonitor); err != nil {
			return nil, fmt.Errorf("update appmonitor '%s/%s' failed", appMonitor.GetNamespace(), appMonitor.GetName())
		}
	}

	return &k8stypes.NamespacedName{
		Namespace: appMonitor.GetNamespace(),
		Name:      appMonitor.GetName(),
	}, nil
}

// do Delete App Monitor
func (h *HttpServerClient) doDeleteAppMonitor(bizID, scenario string) error {
	appMonitorList := &monitorextensionv1.AppMonitorList{}
	selector, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(map[string]string{
		monitorextensionv1.LabelKeyForBizID:        bizID,
		monitorextensionv1.LabelKeyForScenarioName: scenario,
	}))
	if err != nil {
		return fmt.Errorf("build selector failed, err: %w", err)
	}
	err = h.Mgr.GetClient().List(context.Background(), appMonitorList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		return fmt.Errorf("read api-server failed, err: %s", err)
	}

	if len(appMonitorList.Items) == 0 {
		return nil
	}
	if len(appMonitorList.Items) > 1 {
		return fmt.Errorf("unknown error, multi scenario found")
	}

	appMonitor := appMonitorList.Items[0]

	if err = h.Mgr.GetClient().Delete(context.Background(), &appMonitor); err != nil {
		return fmt.Errorf("delete appmonitor'%s/%s' failed, err: %w", appMonitor.GetNamespace(),
			appMonitor.GetName(), err)
	}

	blog.Infof("delete AppMonitor '%s/%s' by http call", appMonitor.Namespace, appMonitor.Name)

	return nil
}
