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

package webhookserver

import (
	"context"
	"testing"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func makeHostNetPortPool(name, ns string, start, end, segLen uint32) *networkextensionv1.HostNetPortPool {
	return &networkextensionv1.HostNetPortPool{
		ObjectMeta: k8smetav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: networkextensionv1.HostNetPortPoolSpec{
			StartPort:     start,
			EndPort:       end,
			SegmentLength: segLen,
		},
	}
}

// TestValidateHostNetPortPool covers field validation and cross-pool range overlap detection.
func TestValidateHostNetPortPool(t *testing.T) {
	newScheme := runtime.NewScheme()
	newScheme.AddKnownTypes(
		networkextensionv1.GroupVersion,
		&networkextensionv1.HostNetPortPool{},
		&networkextensionv1.HostNetPortPoolList{},
	)
	cli := k8sfake.NewFakeClientWithScheme(newScheme)
	server := &Server{k8sClient: cli}

	// existing pool occupies 30000-30100 in ns1.
	existing := makeHostNetPortPool("existing", "ns1", 30000, 30100, 10)
	if err := cli.Create(context.Background(), existing); err != nil {
		t.Fatalf("create existing pool failed: %v", err)
	}

	testCases := []struct {
		title  string
		pool   *networkextensionv1.HostNetPortPool
		hasErr bool
	}{
		{
			title:  "start not less than end",
			pool:   makeHostNetPortPool("bad", "ns1", 30100, 30100, 10),
			hasErr: true,
		},
		{
			title:  "segment length zero",
			pool:   makeHostNetPortPool("bad", "ns1", 30000, 30100, 0),
			hasErr: true,
		},
		{
			title:  "range smaller than segment length",
			pool:   makeHostNetPortPool("bad", "ns1", 30000, 30005, 10),
			hasErr: true,
		},
		{
			title:  "non overlapping pool is allowed",
			pool:   makeHostNetPortPool("new", "ns1", 30100, 30200, 10),
			hasErr: false,
		},
		{
			title:  "overlapping pool same namespace rejected",
			pool:   makeHostNetPortPool("new", "ns1", 30050, 30150, 10),
			hasErr: true,
		},
		{
			title:  "overlapping pool different namespace rejected",
			pool:   makeHostNetPortPool("new", "ns2", 30050, 30150, 10),
			hasErr: true,
		},
		{
			title:  "updating existing pool does not conflict with itself",
			pool:   makeHostNetPortPool("existing", "ns1", 30000, 30120, 10),
			hasErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			err := server.validateHostNetPortPool(tc.pool)
			if tc.hasErr && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tc.hasErr && err != nil {
				t.Fatalf("expected no error but got %v", err)
			}
		})
	}
}
