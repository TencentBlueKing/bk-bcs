/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package webhookserver

import (
	"context"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func getTestExistedPortPools() []networkextensionv1.PortPool {
	return []networkextensionv1.PortPool{
		{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name:      "pool1",
				Namespace: "ns1",
			},
			Spec: networkextensionv1.PortPoolSpec{
				PoolItems: []*networkextensionv1.PortPoolItem{
					{
						ItemName:        "item1",
						LoadBalancerIDs: []string{"ap-shanghai:lb-0011", "lb-0012", "lb-0013"},
						StartPort:       30000,
						EndPort:         30100,
						SegmentLength:   5,
					},
					{
						ItemName:        "item2",
						LoadBalancerIDs: []string{"lb-0021", "ap-nanjing:lb-0022", "lb-0023"},
						StartPort:       30000,
						EndPort:         30100,
					},
				},
			},
		},
		{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name:      "pool2",
				Namespace: "ns1",
			},
			Spec: networkextensionv1.PortPoolSpec{
				PoolItems: []*networkextensionv1.PortPoolItem{
					{
						ItemName:        "item1",
						LoadBalancerIDs: []string{"ap-shanghai:lb-0031", "lb-0032", "lb-0033"},
						StartPort:       30000,
						EndPort:         30100,
					},
					{
						ItemName:        "item2",
						LoadBalancerIDs: []string{"lb-0031", "ap-nanjing:lb-0032", "lb-0033"},
						StartPort:       30000,
						EndPort:         30100,
						SegmentLength:   2,
					},
				},
			},
		},
	}
}

func getTestExistedListeners() []networkextensionv1.Listener {
	return []networkextensionv1.Listener{
		{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name:      "lb1-20000",
				Namespace: "ns2",
			},
			Spec: networkextensionv1.ListenerSpec{
				LoadbalancerID: "lb1",
				Port:           20000,
				Protocol:       "TCP",
			},
		},
		{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name:      "lb2-20000",
				Namespace: "ns2",
				Labels: map[string]string{
					common.GetPortPoolListenerLabelKey("pool1234", "item1234"): networkextensionv1.LabelValueForPortPoolItemName,
				},
			},
			Spec: networkextensionv1.ListenerSpec{
				LoadbalancerID: "lb2",
				Port:           20000,
				Protocol:       "TCP",
			},
		},
		{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name:      "lb1-30000",
				Namespace: "ns1",
			},
			Spec: networkextensionv1.ListenerSpec{
				LoadbalancerID: "lb1",
				Port:           30000,
				Protocol:       "TCP",
			},
		},
		{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name:      "lb1-40000-41000",
				Namespace: "ns1",
			},
			Spec: networkextensionv1.ListenerSpec{
				LoadbalancerID: "lb1",
				Port:           40000,
				EndPort:        41000,
				Protocol:       "TCP",
			},
		},
	}
}

// TestValidatePortPool test validate port pool
func TestValidatePortPool(t *testing.T) {
	newScheme := runtime.NewScheme()
	newScheme.AddKnownTypes(
		networkextensionv1.GroupVersion,
		&networkextensionv1.PortPool{},
		&networkextensionv1.PortPoolList{},
		&networkextensionv1.Listener{},
		&networkextensionv1.ListenerList{},
	)
	cli := k8sfake.NewFakeClientWithScheme(newScheme)
	server := &Server{
		k8sClient: cli,
	}
	for _, pool := range getTestExistedPortPools() {
		if err := cli.Create(context.Background(), &pool); err != nil {
			t.Fatalf("create %v failed, err %s", pool, err.Error())
		}
	}
	for _, li := range getTestExistedListeners() {
		if err := cli.Create(context.Background(), &li); err != nil {
			t.Fatalf("create %v failed, err %s", li, err.Error())
		}
	}

	testCases := []struct {
		title   string
		newPool *networkextensionv1.PortPool
		hasErr  bool
	}{
		{
			title:   "nornal test",
			newPool: &networkextensionv1.PortPool{},
			hasErr:  false,
		},
		{
			title: "conflicts lbids in new port pool",
			newPool: &networkextensionv1.PortPool{
				Spec: networkextensionv1.PortPoolSpec{
					PoolItems: []*networkextensionv1.PortPoolItem{
						{
							ItemName:        "item1",
							LoadBalancerIDs: []string{"lb1", "lb1"},
							StartPort:       30000,
							EndPort:         31000,
						},
					},
				},
			},
			hasErr: true,
		},
		{
			title: "conflicts lbids in new port pool",
			newPool: &networkextensionv1.PortPool{
				Spec: networkextensionv1.PortPoolSpec{
					PoolItems: []*networkextensionv1.PortPoolItem{
						{
							ItemName:        "item1",
							LoadBalancerIDs: []string{"lb1"},
							StartPort:       30000,
							EndPort:         31000,
						},
						{
							ItemName:        "item2",
							LoadBalancerIDs: []string{"lb1"},
							StartPort:       30000,
							EndPort:         31000,
						},
					},
				},
			},
			hasErr: true,
		},
		{
			title: "conflicts item name in new port pool",
			newPool: &networkextensionv1.PortPool{
				Spec: networkextensionv1.PortPoolSpec{
					PoolItems: []*networkextensionv1.PortPoolItem{
						{
							ItemName:        "item1",
							LoadBalancerIDs: []string{"lb1"},
							StartPort:       30000,
							EndPort:         31000,
						},
						{
							ItemName:        "item1",
							LoadBalancerIDs: []string{"lb2"},
							StartPort:       30000,
							EndPort:         31000,
						},
					},
				},
			},
			hasErr: true,
		},
		{
			title: "normal delete item in pool",
			newPool: &networkextensionv1.PortPool{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "pool1",
					Namespace: "ns1",
				},
				Spec: networkextensionv1.PortPoolSpec{
					PoolItems: []*networkextensionv1.PortPoolItem{
						{
							ItemName:        "item1",
							LoadBalancerIDs: []string{"ap-shanghai:lb-0011", "lb-0012", "lb-0013"},
							StartPort:       30000,
							EndPort:         30100,
							SegmentLength:   5,
						},
					},
				},
			},
			hasErr: false,
		},
		{
			title: "error change item in pool",
			newPool: &networkextensionv1.PortPool{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "pool1",
					Namespace: "ns1",
				},
				Spec: networkextensionv1.PortPoolSpec{
					PoolItems: []*networkextensionv1.PortPoolItem{
						{
							ItemName:        "item1",
							LoadBalancerIDs: []string{"ap-shanghai:lb-0011", "lb-99", "lb-0013"},
							StartPort:       30000,
							EndPort:         30100,
							SegmentLength:   2,
						},
						{
							ItemName:        "item2",
							LoadBalancerIDs: []string{"lb-0021", "ap-nanjing:lb-0022", "lb-0023"},
							StartPort:       30000,
							EndPort:         30100,
						},
					},
				},
			},
			hasErr: true,
		},
		{
			title: "same lb id in new pool",
			newPool: &networkextensionv1.PortPool{
				Spec: networkextensionv1.PortPoolSpec{
					PoolItems: []*networkextensionv1.PortPoolItem{
						{
							ItemName:        "item1",
							LoadBalancerIDs: []string{"lb2"},
							StartPort:       30000,
							EndPort:         30999,
						},
						{
							ItemName:        "item2",
							LoadBalancerIDs: []string{"lb2"},
							StartPort:       31000,
							EndPort:         31999,
						},
					},
				},
			},
			hasErr: false,
		},
		{
			title: "same lb id and conflict port",
			newPool: &networkextensionv1.PortPool{
				Spec: networkextensionv1.PortPoolSpec{
					PoolItems: []*networkextensionv1.PortPoolItem{
						{
							ItemName:        "item1",
							LoadBalancerIDs: []string{"lb2"},
							StartPort:       30000,
							EndPort:         31001,
						},
						{
							ItemName:        "item2",
							LoadBalancerIDs: []string{"lb2"},
							StartPort:       31000,
							EndPort:         32000,
						},
					},
				},
			},
			hasErr: true,
		},
		{
			title: "conflicts port with other listener",
			newPool: &networkextensionv1.PortPool{
				Spec: networkextensionv1.PortPoolSpec{
					PoolItems: []*networkextensionv1.PortPoolItem{
						{
							ItemName:        "item1",
							LoadBalancerIDs: []string{"lb1"},
							StartPort:       20000,
							EndPort:         20100,
						},
					},
				},
			},
			hasErr: true,
		},
		{
			title: "exists listener with itselft",
			newPool: &networkextensionv1.PortPool{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "pool1234",
					Namespace: "ns2",
				},
				Spec: networkextensionv1.PortPoolSpec{
					PoolItems: []*networkextensionv1.PortPoolItem{
						{
							ItemName:        "item1234",
							LoadBalancerIDs: []string{"lb2"},
							StartPort:       20000,
							EndPort:         20001,
						},
					},
				},
			},
			hasErr: false,
		},
		{
			title: "conflict with other pool",
			newPool: &networkextensionv1.PortPool{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "pool2",
					Namespace: "ns2",
				},
				Spec: networkextensionv1.PortPoolSpec{
					PoolItems: []*networkextensionv1.PortPoolItem{
						{
							ItemName:        "item1",
							LoadBalancerIDs: []string{"ap-shanghai:lb-0031", "lb-0032", "lb-0033"},
							StartPort:       30000,
							EndPort:         30100,
						},
					},
				},
			},
			hasErr: true,
		},
	}
	for _, test := range testCases {
		t.Run(test.title, func(t *testing.T) {
			if err := server.validatePortPool(test.newPool); err != nil {
				if !test.hasErr {
					t.Fatalf("expect no err, but get err %s", err.Error())
				}
			} else {
				if test.hasErr {
					t.Fatalf("expect err, but get no err")
				}
			}
		})
	}
}
