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

package processor

import (
	ingressv1 "bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/clb/v1"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/common"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/model"
	svcclient "bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient"
	"bk-bcs/bcs-services/bcs-gw-controller/pkg/gw"
	"encoding/json"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8scache "k8s.io/client-go/tools/cache"
)

type MockServiceClient struct {
	Data map[string]*svcclient.AppService
}

func NewMockServiceClient() *MockServiceClient {
	data := make(map[string]*svcclient.AppService)
	return &MockServiceClient{
		Data: data,
	}
}

func (msc *MockServiceClient) AddAppService(appSvc *svcclient.AppService) {
	msc.Data[appSvc.Namespace+"/"+appSvc.Name] = appSvc
}

func (msc *MockServiceClient) GetAppService(ns, name string) (*svcclient.AppService, error) {
	return msc.Data[ns+"/"+name], nil
}

func (msc *MockServiceClient) ListAppService(labels map[string]string) ([]*svcclient.AppService, error) {
	var retList []*svcclient.AppService
	for _, appsvc := range msc.Data {
		labelMatch := true
		for key, value := range labels {
			existedLabelValue, ok := appsvc.Labels[key]
			if !ok {
				labelMatch = false
				break
			}
			if existedLabelValue != value {
				labelMatch = false
				break
			}
		}
		if !labelMatch {
			continue
		}
		retList = append(retList, appsvc)
	}
	return retList, nil
}

func (msc *MockServiceClient) ListAppServiceFromStatefulSet(ns, name string) ([]*svcclient.AppService, error) {
	return nil, nil
}

func (msc *MockServiceClient) Close() {}

type MockIngressClient struct {
	Data map[string]*ingressv1.ClbIngress
}

func NewMockIngressClient() *MockIngressClient {
	data := make(map[string]*ingressv1.ClbIngress)
	return &MockIngressClient{
		Data: data,
	}
}

func (mic *MockIngressClient) AddIngress(ingress *ingressv1.ClbIngress) {
	mic.Data[ingress.Name] = ingress
}

func (mic *MockIngressClient) AddIngressHandler(handler model.EventHandler) {}

func (mic *MockIngressClient) ListIngresses() ([]*ingressv1.ClbIngress, error) {
	var retList []*ingressv1.ClbIngress
	for _, ingress := range mic.Data {
		retList = append(retList, ingress)
	}
	return retList, nil
}

func (mic *MockIngressClient) GetIngress(name string) (*ingressv1.ClbIngress, error) {
	return mic.Data[name], nil
}

func (mic *MockIngressClient) SetIngress(*ingressv1.ClbIngress) error {
	return nil
}

func TestUpdaterAdd(t *testing.T) {
	updater := &GWUpdater{
		cache: k8scache.NewStore(GwServiceKeyFunc),
	}
	msc := NewMockServiceClient()
	updater.SetServiceClient(msc)
	updater.SetOption(&Option{
		Port:          18080,
		BackendIPType: common.BackendIPTypeUnderlay,
		Cluster:       "cluster1",
		GwBizID:       "测试1",
		ServiceLabel: map[string]string{
			"gw.bkbcs.tencent.com": "gw",
		},
		DomainLabelKey:    "domain.gw.bkbcs.tencent.com",
		ProxyPortLabelKey: "proxyport.gw.bkbcs.tencent.com",
		PortLabelKey:      "port.gw.bkbcs.tencent.com",
		PathLabelKey:      "path.gw.bkbcs.tencent.com",
	})
	gwClient := &gw.MockClient{}
	updater.SetGWClient(gwClient)
	msc.AddAppService(&svcclient.AppService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc1",
			Namespace: "ns1",
			Labels: map[string]string{
				"gw.bkbcs.tencent.com":           "gw",
				"domain.gw.bkbcs.tencent.com":    "www.test1.com",
				"proxyport.gw.bkbcs.tencent.com": "28080",
				"port.gw.bkbcs.tencent.com":      "18080",
				"path.gw.bkbcs.tencent.com":      "test.test1",
			},
		},
		ServicePorts: []svcclient.ServicePort{
			{
				Name:        "port-1",
				Protocol:    "HTTPS",
				Domain:      "www.test1.com",
				Path:        "/test/test1",
				ServicePort: 18080,
				TargetPort:  8080,
			},
		},
		Nodes: []svcclient.AppNode{
			{
				NodeIP: "127.0.0.11",
				Ports: []svcclient.NodePort{
					{
						Name:      "port-1",
						Protocol:  "HTTPS",
						NodePort:  8080,
						ProxyPort: 0,
					},
				},
			},
			{
				NodeIP: "127.0.0.12",
				Ports: []svcclient.NodePort{
					{
						Name:      "port-1",
						Protocol:  "HTTPS",
						NodePort:  8080,
						ProxyPort: 0,
					},
				},
			},
		},
	})

	msc.AddAppService(&svcclient.AppService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc2",
			Namespace: "ns2",
			Labels: map[string]string{
				"gw.bkbcs.tencent.com":           "gw",
				"proxyport.gw.bkbcs.tencent.com": "38080",
				"port.gw.bkbcs.tencent.com":      "18080",
				"path.gw.bkbcs.tencent.com":      "test.test1",
			},
		},
		ServicePorts: []svcclient.ServicePort{
			{
				Name:        "port-1",
				Protocol:    "HTTPS",
				Domain:      "www.test2.com",
				Path:        "/test/test1",
				ServicePort: 18080,
				TargetPort:  8080,
			},
		},
		Nodes: []svcclient.AppNode{
			{
				NodeIP: "127.0.0.11",
				Ports: []svcclient.NodePort{
					{
						Name:      "port-1",
						Protocol:  "HTTPS",
						NodePort:  8080,
						ProxyPort: 0,
					},
				},
			},
			{
				NodeIP: "127.0.0.12",
				Ports: []svcclient.NodePort{
					{
						Name:      "port-1",
						Protocol:  "HTTPS",
						NodePort:  8080,
						ProxyPort: 0,
					},
				},
			},
		},
	})

	err := updater.Update()
	if err != nil {
		t.Errorf(err.Error())
	}

	svcs, _ := gwClient.List()
	refSvcs := []*gw.Service{
		{
			BizID:                 "测试1",
			Domain:                "www.test1.com",
			VPort:                 28080,
			Type:                  gw.ProtocolHTTPS,
			SSLEnable:             true,
			SSLVerifyClientEnable: false,
			LocationList: []*gw.Location{
				{
					URL: "/test/test1",
					RSList: []*gw.RealServer{
						{
							IP:     "127.0.0.11",
							Port:   8080,
							Weight: 100,
						},
						{
							IP:     "127.0.0.12",
							Port:   8080,
							Weight: 100,
						},
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(svcs, refSvcs) {
		svcsBytes, _ := json.Marshal(svcs)
		refSvcsBytes, _ := json.Marshal(refSvcs)
		t.Errorf("%s should be equal to %s", string(svcsBytes), string(refSvcsBytes))
	}
}
