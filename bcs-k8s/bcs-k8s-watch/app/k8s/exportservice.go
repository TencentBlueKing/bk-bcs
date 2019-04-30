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

package k8s

import (
	"fmt"
	"sort"
	"strings"

	//"k8s.io/client-go/pkg/api/v1"
	//"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output/action"

	commtypes "bk-bcs/bcs-common/common/types"
	lbtypes "bk-bcs/bcs-common/pkg/loadbalance/v2"
)

const ResourceKind = "ExportService"

type ExportServiceInterface interface {
	SyncIngress(item interface{})
}

// ExportServiceController as esc, ExportService as es
type ExportServiceController struct {
	lister    *StoreLister
	writer    *output.Writer
	clusterID string
}

type ExportServiceWithAction struct {
	exportService *lbtypes.ExportService
	action        string
}

func NewExportServiceController() *ExportServiceController {
	es := &ExportServiceController{}
	lister := &StoreLister{}
	es.lister = lister
	return es
}

func (esc *ExportServiceController) SyncIngress(originEvent OriginEvent) {

	exportServicesWithAction := esc.createExportServiceListWithAction(originEvent)
	for _, exportServiceWithAction := range exportServicesWithAction {
		syncData := &action.SyncData{
			Kind:      ResourceKind,
			Namespace: exportServiceWithAction.exportService.Namespace,
			Name:      exportServiceWithAction.exportService.ServiceName,
			Action:    exportServiceWithAction.action,
			Data:      exportServiceWithAction.exportService,
		}

		glog.V(2).Infof(
			"Will Sync ExportService: %s, Action: %s", exportServiceWithAction.exportService.ServiceName, exportServiceWithAction.action)
		esc.writer.Sync(syncData)
	}
}

func (esc *ExportServiceController) createExportServiceListWithAction(originEvent OriginEvent) (esList []*ExportServiceWithAction) {

	var ingressList []*v1beta1.Ingress
	// the adding and deleting action of Ingress will map to exportService's same action, others will map to 'Update' action
	var esAction = "Update"
	if originEvent.ResourceType == "Ingress" {
		// assume the delete action has removed the ingress
		if originEvent.Action == "Delete" {
			exportServiceWithAction := &ExportServiceWithAction{
				exportService: &lbtypes.ExportService{
					Namespace:   originEvent.Namespace,
					ServiceName: originEvent.ResourceName,
				},
				action: "Delete",
			}
			esList = append(esList, exportServiceWithAction)
			return
		}
		item, exist, err := esc.lister.Ingress.GetByKey(fmt.Sprintf("%s/%s", originEvent.Namespace, originEvent.ResourceName))
		if err != nil {
			glog.Errorf("get ingress resource failed: %s", err.Error())
			return
		}
		if !exist {
			glog.Errorf("get no ingress, not as expected")
			return
		}

		if targetIngress, ok := item.(*v1beta1.Ingress); ok {
			ingressList = append(ingressList, targetIngress)
			esAction = originEvent.Action
		}


	} else {
		glog.Infof("got ingress related resource changed, have to update all ingress")
		ings := esc.lister.Ingress.List()
		sort.SliceStable(ings, func(i, j int) bool {
			ir := ings[i].(*v1beta1.Ingress).ResourceVersion
			jr := ings[j].(*v1beta1.Ingress).ResourceVersion
			return ir < jr
		})

		for _, ingIf := range ings {
			ingressList = append(ingressList, ingIf.(*v1beta1.Ingress))
		}
	}

	for _, ingressInstance := range ingressList {
		exportService := esc.createExportService(ingressInstance)
		exportServiceWithAction := &ExportServiceWithAction{
			exportService: exportService,
			action:        esAction,
		}
		esList = append(esList, exportServiceWithAction)
	}

	return
}

func (esc *ExportServiceController) createExportService(ingressInstance *v1beta1.Ingress) *lbtypes.ExportService {
	exportService := &lbtypes.ExportService{
		ObjectMeta: commtypes.ObjectMeta{
			Name:        ingressInstance.Name,
			NameSpace:   ingressInstance.Namespace,
			Labels:      ingressInstance.Labels,
			Annotations: ingressInstance.Annotations,
		},
		Balance:       "roundrobin",
		SSLCert:       false,
		MaxConn:       50000,
		Cluster:       esc.clusterID,
		ServiceName:   ingressInstance.Name,
		ServiceWeight: map[string]int{},
		BCSGroup:      []string{"external"},
		Namespace:     ingressInstance.Namespace,
	}

	var exportPorts []lbtypes.ExportPort

	// for http
	for _, rule := range ingressInstance.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			serviceInstance, endpointsInstance := esc.getServiceAndEndPoints(ingressInstance.Namespace, path.Backend.ServiceName)
			if serviceInstance == nil || endpointsInstance == nil {
				continue
			}
			exportPorts = append(exportPorts, esc.getExportPort(path.Path, rule.Host, serviceInstance, endpointsInstance)...)
		}
	}

	// for tcp
	if ingressInstance.Spec.Backend != nil {
		serviceInstance, endpointsInstance := esc.getServiceAndEndPoints(ingressInstance.Namespace, ingressInstance.Spec.Backend.ServiceName)
		if serviceInstance == nil || endpointsInstance == nil {
			exportService.ServicePort = exportPorts
			return exportService
		}
		exportPorts = append(exportPorts, esc.getExportPort("", "", serviceInstance, endpointsInstance)...)
	}

	glog.V(2).Infof("serviceName: %s, servicePort: %+v", exportService.ServiceName, exportPorts)
	exportService.ServicePort = exportPorts
	return exportService
}

func (esc *ExportServiceController) getServiceAndEndPoints(namespace string, serviceName string) (service *v1.Service, endPoint *v1.Endpoints) {
	item, exist, err := esc.lister.Service.GetByKey(
		fmt.Sprintf("%s/%s", namespace, serviceName))
	if err != nil {
		glog.Errorf("namespace(%s) get service(%s) failed", namespace, serviceName)
		return nil, nil
	}
	if !exist {
		// ignore the non-service ingress, so stop printing
		glog.Warningf("namespace(%s) get no service(%s)", namespace, serviceName)
		return nil, nil
	}

	service, ok := item.(*v1.Service)
	if !ok {
		glog.Errorf("serviceInstance(%s) assert type v1.Service failed", serviceName)
		return nil, nil
	}

	item, exist, err = esc.lister.Endpoint.GetByKey(
		fmt.Sprintf("%s/%s", namespace, serviceName))
	if err != nil {
		glog.Errorf("ingress(%s) get endpoint(%s) failed", namespace, serviceName)
		return nil, nil
	}
	if !exist {
		glog.Warningf("ingress(%s) get no endpoint(%s)", namespace, serviceName)
		return nil, nil
	}

	endPoint, ok = item.(*v1.Endpoints)
	if !ok {
		glog.Errorf("endpoints(%s) assert type v1.EndPoints failed", serviceName)
		return nil, nil
	}

	return
}

// copy from bcs-core
func (esc *ExportServiceController) getExportPort(path string, bcsVhost string, svc *v1.Service, endpoint *v1.Endpoints) (exportPorts []lbtypes.ExportPort) {
	for _, end := range endpoint.Subsets {
		for _, p := range end.Ports {
			// get svc port from svc
			var svcPort int
			var svcPortname string
			var protocol string
		outter:
			for _, svcOuterPort := range svc.Spec.Ports {
				switch svcOuterPort.TargetPort.Type {
				case intstr.String:
					if p.Name == svcOuterPort.TargetPort.StrVal {
						svcPortname = p.Name
						svcPort = int(svcOuterPort.Port)
						protocol = strings.ToLower(string(svcOuterPort.Protocol))
						break outter
					}
				case intstr.Int:
					if p.Port == svcOuterPort.TargetPort.IntVal {
						svcPortname = p.Name
						svcPort = int(svcOuterPort.Port)
						protocol = strings.ToLower(string(svcOuterPort.Protocol))
						break outter
					}
				}
			}

			var backends lbtypes.BackendList
			for _, addr := range end.Addresses {
				var back lbtypes.Backend
				back.TargetPort = int(p.Port)
				back.TargetIP = addr.IP
				backends = append(backends, back)
			}

			exportPort := lbtypes.ExportPort{
				Name:        svcPortname,
				BCSVHost:    bcsVhost,
				Path:        path,
				Protocol:    protocol,
				ServicePort: svcPort,
				Backends:    backends,
			}
			if len(exportPort.BCSVHost) != 0 {
				exportPort.Protocol = "http"
			}
			exportPorts = append(exportPorts, exportPort)
		}
	}
	return
}
