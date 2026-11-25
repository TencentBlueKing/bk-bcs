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

package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cloudaccount"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"

	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

var errNotFound = errors.New("not found")

type stubStoreModel struct {
	store.ClusterManagerModel
	clusters   map[string]*types.Cluster
	nodeGroups map[string]*types.NodeGroup
	tasks      map[string]*types.Task
	accounts   map[string]*types.CloudAccount
}

func (s *stubStoreModel) GetCluster(_ context.Context, clusterID string) (*types.Cluster, error) {
	if c, ok := s.clusters[clusterID]; ok {
		return c, nil
	}
	return nil, errNotFound
}

func (s *stubStoreModel) GetNodeGroup(_ context.Context, groupID string) (*types.NodeGroup, error) {
	if np, ok := s.nodeGroups[groupID]; ok {
		return np, nil
	}
	return nil, errNotFound
}

func (s *stubStoreModel) GetTask(_ context.Context, taskID string) (*types.Task, error) {
	if t, ok := s.tasks[taskID]; ok {
		return t, nil
	}
	return nil, errNotFound
}

func (s *stubStoreModel) GetCloudAccount(_ context.Context, cloudID, accountID string, _ bool) (
	*types.CloudAccount, error) {
	if a, ok := s.accounts[cloudID+"/"+accountID]; ok {
		return a, nil
	}
	return nil, errNotFound
}

func newStubStoreModel() *stubStoreModel {
	return &stubStoreModel{
		clusters: map[string]*types.Cluster{
			"cls-a":      {ClusterID: "cls-a", ProjectID: "proj-a"},
			"cls-b":      {ClusterID: "cls-b", ProjectID: "proj-b"},
			"cls-shared": {ClusterID: "cls-shared", ProjectID: "proj-b", IsShared: true},
		},
		nodeGroups: map[string]*types.NodeGroup{
			"np-b": {NodeGroupID: "np-b", ClusterID: "cls-b", ProjectID: "proj-b"},
		},
		tasks: map[string]*types.Task{
			"task-b": {TaskID: "task-b", ClusterID: "cls-b", ProjectID: "proj-b"},
		},
		accounts: map[string]*types.CloudAccount{
			"tencentCloud/acct-b":    {CloudID: "tencentCloud", AccountID: "acct-b", ProjectID: "proj-b"},
			"tencentCloud/acct-none": {CloudID: "tencentCloud", AccountID: "acct-none"},
		},
	}
}

func TestCheckResourceID(t *testing.T) {
	tests := []struct {
		name        string
		action      string
		in          resourceID
		wantErr     bool
		wantProject string
		wantCluster string
	}{
		{
			name:        "cluster project filled from store",
			action:      cluster.CanManageClusterOperation,
			in:          resourceID{ClusterID: "cls-b"},
			wantProject: "proj-b",
			wantCluster: "cls-b",
		},
		{
			name:        "cluster project matches",
			action:      cluster.CanManageClusterOperation,
			in:          resourceID{ClusterID: "cls-b", ProjectID: "proj-b"},
			wantProject: "proj-b",
			wantCluster: "cls-b",
		},
		{
			name:    "cluster project mismatch",
			action:  cluster.CanManageClusterOperation,
			in:      resourceID{ClusterID: "cls-b", ProjectID: "proj-a"},
			wantErr: true,
		},
		{
			name:    "cluster project mismatch on view",
			action:  cluster.CanViewClusterOperation,
			in:      resourceID{ClusterID: "cls-b", ProjectID: "proj-a"},
			wantErr: true,
		},
		{
			name:        "shared cluster viewed by other project",
			action:      cluster.CanViewClusterOperation,
			in:          resourceID{ClusterID: "cls-shared", ProjectID: "proj-a"},
			wantProject: "proj-a",
			wantCluster: "cls-shared",
		},
		{
			name:    "shared cluster managed by other project",
			action:  cluster.CanManageClusterOperation,
			in:      resourceID{ClusterID: "cls-shared", ProjectID: "proj-a"},
			wantErr: true,
		},
		{
			name:    "cluster not found",
			action:  cluster.CanViewClusterOperation,
			in:      resourceID{ClusterID: "cls-x", ProjectID: "proj-a"},
			wantErr: true,
		},
		{
			name:        "server key as cluster",
			action:      cluster.CanManageClusterOperation,
			in:          resourceID{ServerKey: "cls-b"},
			wantProject: "proj-b",
			wantCluster: "cls-b",
		},
		{
			name:        "nodegroup fills cluster and project",
			action:      cluster.CanManageClusterOperation,
			in:          resourceID{NodeGroupID: "np-b"},
			wantProject: "proj-b",
			wantCluster: "cls-b",
		},
		{
			name:    "nodegroup cluster mismatch",
			action:  cluster.CanManageClusterOperation,
			in:      resourceID{NodeGroupID: "np-b", ClusterID: "cls-a"},
			wantErr: true,
		},
		{
			name:    "nodegroup project mismatch",
			action:  cluster.CanManageClusterOperation,
			in:      resourceID{NodeGroupID: "np-b", ProjectID: "proj-a"},
			wantErr: true,
		},
		{
			name:        "task fills cluster and project",
			action:      cluster.CanManageClusterOperation,
			in:          resourceID{TaskID: "task-b"},
			wantProject: "proj-b",
			wantCluster: "cls-b",
		},
		{
			name:    "task cluster mismatch",
			action:  cluster.CanManageClusterOperation,
			in:      resourceID{TaskID: "task-b", ClusterID: "cls-a", ProjectID: "proj-a"},
			wantErr: true,
		},
		{
			name:        "cloud account project filled from store",
			action:      cloudaccount.CanUseCloudAccountOperation,
			in:          resourceID{CloudID: "tencentCloud", AccountID: "acct-b"},
			wantProject: "proj-b",
		},
		{
			name:    "cloud account project mismatch",
			action:  cloudaccount.CanManageCloudAccountOperation,
			in:      resourceID{CloudID: "tencentCloud", AccountID: "acct-b", ProjectID: "proj-a"},
			wantErr: true,
		},
		{
			name:        "cloud account without project keeps request project",
			action:      cloudaccount.CanUseCloudAccountOperation,
			in:          resourceID{CloudID: "tencentCloud", AccountID: "acct-none", ProjectID: "proj-a"},
			wantProject: "proj-a",
		},
		{
			name:        "create cluster does not look up cluster",
			action:      cluster.CanCreateClusterOperation,
			in:          resourceID{ClusterID: "cls-new", ProjectID: "proj-a"},
			wantProject: "proj-a",
			wantCluster: "cls-new",
		},
		{
			name:        "project action keeps request project",
			action:      project.CanViewProjectOperation,
			in:          resourceID{ClusterID: "cls-b", ProjectID: "proj-a"},
			wantProject: "proj-a",
			wantCluster: "cls-b",
		},
		{
			name:        "project action fills project from cloud account",
			action:      project.CanViewProjectOperation,
			in:          resourceID{CloudID: "tencentCloud", AccountID: "acct-b"},
			wantProject: "proj-b",
		},
		{
			name:    "project action cloud account project mismatch",
			action:  project.CanViewProjectOperation,
			in:      resourceID{CloudID: "tencentCloud", AccountID: "acct-b", ProjectID: "proj-a"},
			wantErr: true,
		},
	}

	model := newStubStoreModel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.in
			err := resolveResourceID(context.Background(), model, &r)
			if err == nil {
				err = verifyResourceOwner(context.Background(), model, &r, tt.action)
			}
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got resource %+v", r)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if r.ProjectID != tt.wantProject {
				t.Errorf("projectID = %q, want %q", r.ProjectID, tt.wantProject)
			}
			if r.ClusterID != tt.wantCluster {
				t.Errorf("clusterID = %q, want %q", r.ClusterID, tt.wantCluster)
			}
		})
	}
}

func TestResolveKeepsRequestValues(t *testing.T) {
	r := resourceID{NodeGroupID: "np-b", ProjectID: "proj-a"}
	if err := resolveResourceID(context.Background(), newStubStoreModel(), &r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ClusterID != "cls-b" || r.ProjectID != "proj-a" {
		t.Errorf("got cluster %q project %q, want cls-b proj-a", r.ClusterID, r.ProjectID)
	}
}
