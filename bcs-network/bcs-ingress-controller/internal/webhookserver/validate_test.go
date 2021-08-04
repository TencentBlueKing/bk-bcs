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

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"

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
						EndPort:         40000,
						SegmentLength:   2,
					},
					{
						ItemName:        "item2",
						LoadBalancerIDs: []string{"lb-0021", "ap-nanjing:lb-0022", "lb-0023"},
						StartPort:       30000,
						EndPort:         40000,
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
						EndPort:         40000,
					},
					{
						ItemName:        "item2",
						LoadBalancerIDs: []string{"lb-0031", "ap-nanjing:lb-0032", "lb-0033"},
						StartPort:       30000,
						EndPort:         40000,
						SegmentLength:   2,
					},
				},
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
							EndPort:         40000,
							SegmentLength:   2,
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
							EndPort:         40000,
							SegmentLength:   2,
						},
						{
							ItemName:        "item2",
							LoadBalancerIDs: []string{"lb-0021", "ap-nanjing:lb-0022", "lb-0023"},
							StartPort:       30000,
							EndPort:         40000,
						},
					},
				},
			},
			hasErr: true,
		},
	}
	for id, test := range testCases {
		t.Logf("test %d: %s", id, test.title)
		if err := server.validatePortPool(test.newPool); err != nil {
			if !test.hasErr {
				t.Fatalf("expect no err, but get err %s", err.Error())
			}
		} else {
			if test.hasErr {
				t.Fatalf("expect err, but get no err")
			}
		}

	}
}
