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

package cluster

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/cluster/requester/mocks"
)

func TestClient_ListClusterNodes(t *testing.T) {
	// nolint
	rawNodeList := []byte("{\"items\":[{\"metadata\":{\"name\":\"1.1.1.1\",\"labels\":{\"beta.kubernetes.io/arch\":\"amd64\"}},\"spec\":{\"podCIDR\":\"1.1.1.1/27\",\"podCIDRs\":[\"1.1.1.1/27\"],\"providerID\":\"qcloud:///330001/ins-d3cwxd6w\"},\"status\":{\"capacity\":{\"cpu\":\"4\",\"ephemeral-storage\":\"103079844Ki\",\"hugepages-1Gi\":\"0\",\"hugepages-2Mi\":\"0\",\"memory\":\"16128872Ki\",\"pods\":\"29\"},\"conditions\":[],\"addresses\":[{\"type\":\"InternalIP\",\"address\":\"1.1.1.1\"},{\"type\":\"Hostname\",\"address\":\"1.1.1.1\"}],\"nodeInfo\":{},\"images\":[]}},{\"metadata\":{\"name\":\"2.2.2.2\",\"labels\":{\"beta.kubernetes.io/arch\":\"amd64\"}},\"status\":{\"capacity\":{\"cpu\":\"8\",\"ephemeral-storage\":\"103079844Ki\",\"hugepages-1Gi\":\"0\",\"hugepages-2Mi\":\"0\",\"memory\":\"64567304Ki\",\"pods\":\"29\"},\"conditions\":[],\"addresses\":[{\"type\":\"InternalIP\",\"address\":\"2.2.2.2\"},{\"type\":\"Hostname\",\"address\":\"2.2.2.2\"}],\"daemonEndpoints\":{\"kubeletEndpoint\":{\"Port\":10250}},\"nodeInfo\":{},\"images\":[]}}]}\n")
	tests := []struct {
		name      string
		clusterID string
		want      int
		wantErr   bool
		on        func(f *MockFields)
	}{
		{
			name:      "normal",
			clusterID: "BCS-K8S-15202",
			want:      2,
			wantErr:   false,
			on: func(f *MockFields) {
				f.requester.On("DoGetRequest",
					fmt.Sprintf("/clusters/%s/api/v1/nodes?labelSelector=node-role.kubernetes.io/master!=true", "BCS-K8S-15202"),
					mock.Anything).Return(rawNodeList, nil)
			},
		},
		{
			name:      "err",
			clusterID: "BCS-K8S-15202",
			want:      0,
			wantErr:   true,
			on: func(f *MockFields) {
				f.requester.On("DoGetRequest",
					fmt.Sprintf("/clusters/%s/api/v1/nodes?labelSelector=node-role.kubernetes.io/master!=true", "BCS-K8S-15202"),
					mock.Anything).
					Return(nil, fmt.Errorf("request err"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRequester := &MockFields{
				requester: mocks.NewRequester(t),
			}
			tt.on(mockRequester)
			opts := &ClusterClientOptions{
				Endpoint: "",
				Token:    "",
				Sender:   mockRequester.requester,
			}
			cmOpts := &ClusterManagerClientOptions{}
			client := NewClient(opts, cmOpts)
			rsp, err := client.ListClusterNodes(tt.clusterID)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, len(rsp))
		})
	}
}

func TestClient_UpdateNodeLabels(t *testing.T) {
	tests := []struct {
		name      string
		clusterID string
		nodeName  string
		label     map[string]interface{}
		wantErr   bool
		on        func(f *MockFields)
	}{
		{
			name:      "normal",
			clusterID: "BCS-K8S-15202",
			nodeName:  "1.1.1.1",
			label:     map[string]interface{}{"test": "test"},
			wantErr:   false,
			on: func(f *MockFields) {
				f.requester.On("DoPatchRequest",
					fmt.Sprintf("/clusters/%s/api/v1/nodes/%s", "BCS-K8S-15202", "1.1.1.1"), mock.Anything, mock.Anything).
					Return([]byte(""), nil)
			},
		},
		{
			name:      "err",
			clusterID: "BCS-K8S-15202",
			nodeName:  "2.2.2.2",
			label:     map[string]interface{}{"test": "test"},
			wantErr:   true,
			on: func(f *MockFields) {
				f.requester.On("DoPatchRequest",
					fmt.Sprintf("/clusters/%s/api/v1/nodes/%s", "BCS-K8S-15202", "2.2.2.2"), mock.Anything, mock.Anything).
					Return([]byte(""), fmt.Errorf("patch err"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRequester := &MockFields{
				requester: mocks.NewRequester(t),
			}
			tt.on(mockRequester)
			opts := &ClusterClientOptions{
				Endpoint: "",
				Token:    "",
				Sender:   mockRequester.requester,
			}
			cmOpts := &ClusterManagerClientOptions{}
			client := NewClient(opts, cmOpts)
			err := client.UpdateNodeMetadata(tt.clusterID, tt.nodeName, tt.label, nil)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestClient_ListNodesByLabel(t *testing.T) {
	// nolint
	rawNodeList := []byte("{\"items\":[{\"metadata\":{\"name\":\"1.1.1.1\",\"labels\":{\"beta.kubernetes.io/arch\":\"amd64\"}},\"spec\":{\"podCIDR\":\"1.1.1.1/27\",\"podCIDRs\":[\"1.1.1.1/27\"],\"providerID\":\"qcloud:///330001/ins-d3cwxd6w\"},\"status\":{\"capacity\":{\"cpu\":\"4\",\"ephemeral-storage\":\"103079844Ki\",\"hugepages-1Gi\":\"0\",\"hugepages-2Mi\":\"0\",\"memory\":\"16128872Ki\",\"pods\":\"29\"},\"conditions\":[],\"addresses\":[{\"type\":\"InternalIP\",\"address\":\"1.1.1.1\"},{\"type\":\"Hostname\",\"address\":\"1.1.1.1\"}],\"nodeInfo\":{},\"images\":[]}},{\"metadata\":{\"name\":\"2.2.2.2\",\"labels\":{\"beta.kubernetes.io/arch\":\"amd64\"}},\"status\":{\"capacity\":{\"cpu\":\"8\",\"ephemeral-storage\":\"103079844Ki\",\"hugepages-1Gi\":\"0\",\"hugepages-2Mi\":\"0\",\"memory\":\"64567304Ki\",\"pods\":\"29\"},\"conditions\":[],\"addresses\":[{\"type\":\"InternalIP\",\"address\":\"2.2.2.2\"},{\"type\":\"Hostname\",\"address\":\"2.2.2.2\"}],\"daemonEndpoints\":{\"kubeletEndpoint\":{\"Port\":10250}},\"nodeInfo\":{},\"images\":[]}}]}\n")
	tests := []struct {
		name      string
		clusterID string
		label     map[string]interface{}
		want      int
		wantErr   bool
		on        func(f *MockFields)
	}{
		{
			name:      "normal",
			clusterID: "BCS-K8S-15202",
			label:     map[string]interface{}{"test": "test"},
			want:      2,
			wantErr:   false,
			on: func(f *MockFields) {
				f.requester.On("DoGetRequest",
					// nolint
					fmt.Sprintf("/clusters/%s/api/v1/nodes?labelSelector=node-role.kubernetes.io/master!=true,test=test", "BCS-K8S-15202"),
					mock.Anything).Return(rawNodeList, nil)
			},
		},
		{
			name:      "err",
			clusterID: "BCS-K8S-15202",
			label:     map[string]interface{}{"test": "test"},
			want:      0,
			wantErr:   true,
			on: func(f *MockFields) {
				f.requester.On("DoGetRequest",
					// nolint
					fmt.Sprintf("/clusters/%s/api/v1/nodes?labelSelector=node-role.kubernetes.io/master!=true,test=test", "BCS-K8S-15202"), mock.Anything).
					Return(nil, fmt.Errorf("request err"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRequester := &MockFields{
				requester: mocks.NewRequester(t),
			}
			tt.on(mockRequester)
			opts := &ClusterClientOptions{
				Endpoint: "",
				Token:    "",
				Sender:   mockRequester.requester,
			}
			cmOpts := &ClusterManagerClientOptions{}
			client := NewClient(opts, cmOpts)
			rsp, err := client.ListNodesByLabel(tt.clusterID, tt.label)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, len(rsp))
		})
	}
}

type MockFields struct {
	requester *mocks.Requester
}
