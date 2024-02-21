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

package cmdb

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	bkcmdbkube "configcenter/src/kube/types"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/option"
)

var (
	// BkcmdbSynchronizerOption is an option for the BkcmdbSynchronizer.
	BkcmdbSynchronizerOption = &option.BkcmdbSynchronizerOption{}
	bkBizID                  = int64(41)
)

func init() {
	//jsonFile, err := os.Open("bcs-bkcmdb-synchronizer-decrypted.json")
	//if err != nil {
	//	panic(err)
	//}
	//defer jsonFile.Close()
	//byteValue, _ := io.ReadAll(jsonFile)
	//json.Unmarshal(byteValue, BkcmdbSynchronizerOption)
}

func getCli() *cmdbClient {
	return NewCmdbClient(&Options{
		AppCode:    os.Getenv("TEST_CMDB_BK_APP_CODE"),
		AppSecret:  os.Getenv("TEST_CMDB_BK_APP_SECRET"),
		BKUserName: os.Getenv("TEST_CMDB_BK_USERNAME"),
		Server:     os.Getenv("TEST_CMDB_SERVER"),
		Debug:      true,
	})
}

// Test_cmdbClient_GetBcsCluster tests the GetBcsCluster method of the cmdbClient.
func Test_cmdbClient_GetBcsCluster(t *testing.T) {
	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.GetBcsClusterRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *[]bkcmdbkube.Cluster
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.GetBcsClusterRequest{
					CommonRequest: client.CommonRequest{
						BKBizID: 41,
						Page: client.Page{
							Limit: 100,
							Start: 0,
						},
						Fields: []string{},
						Filter: &client.PropertyFilter{
							Condition: "AND",
							Rules: []client.Rule{
								{
									Field:    "id",
									Operator: "in",
									Value:    []int64{4583},
								},
							},
						},
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			got, err := c.GetBcsCluster(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBcsCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("GetBcsCluster() got = %v", got)
			clusterids := make([]int64, 0)
			for _, cluster := range *got {
				clusterids = append(clusterids, cluster.ID)
			}

			for _, clusterid := range clusterids {
				fmt.Printf("%d,", clusterid)
			}
		})
	}
}

// Test_cmdbClient_CreateBcsCluster tests the CreateBcsCluster method of the cmdbClient.
func Test_cmdbClient_CreateBcsCluster(t *testing.T) {
	name := "cluster-bcs-99s9556c0"
	schedulingEngine := "k8s"
	uid := "BCS-K8S-910215519833"
	xid := "BCS-K8S-910003553"
	version := "1.19.3"
	networkType := "underlay"
	region := "ap-nanjing"
	vpc := "vpc-xxxx"
	network := []string{"1.1.1.0/21"}
	clusterType := "INDEPENDENT_CLUSTER"
	environment := "uat"

	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.CreateBcsClusterRequest
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantBkClusterID int64
		wantErr         bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.CreateBcsClusterRequest{
					BKBizID:          &bkBizID,
					Name:             &name,
					SchedulingEngine: &schedulingEngine,
					UID:              &uid,
					XID:              &xid,
					Version:          &version,
					NetworkType:      &networkType,
					Region:           &region,
					Vpc:              &vpc,
					Network:          &network,
					Type:             &clusterType,
					Environment:      &environment,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			gotBkClusterID, err := c.CreateBcsCluster(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBcsCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("CreateBcsCluster() gotBkClusterID = %v", gotBkClusterID)
		})
	}
}

// Test_cmdbClient_UpdateBcsCluster tests the UpdateBcsCluster method of the cmdbClient.
func Test_cmdbClient_UpdateBcsCluster(t *testing.T) {
	clusterID := []int64{3603}
	//tmp := "123"
	environment := "debug"

	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.UpdateBcsClusterRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.UpdateBcsClusterRequest{
					BKBizID: &bkBizID,
					IDs:     &clusterID,
					Data: &client.UpdateBcsClusterRequestData{
						//Version:     &tmp,
						//NetworkType: &tmp,
						//Region:      &tmp,
						//Network:     &[]string{"111"},
						Environment: &environment,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			if err := c.UpdateBcsCluster(tt.args.request); (err != nil) != tt.wantErr {
				t.Errorf("UpdateBcsCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_cmdbClient_DeleteBcsCluster tests the DeleteBcsCluster method of the cmdbClient.
func Test_cmdbClient_DeleteBcsCluster(t *testing.T) {
	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.DeleteBcsClusterRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.DeleteBcsClusterRequest{
					BKBizID: &bkBizID,
					IDs:     &[]int64{3602, 3607, 4362, 4374, 4401, 4402, 4403},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			if err := c.DeleteBcsCluster(tt.args.request); (err != nil) != tt.wantErr {
				t.Errorf("DeleteBcsCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_cmdbClient_GetBcsNamespace tests the GetBcsNamespace method of the cmdbClient.
func Test_cmdbClient_GetBcsNamespace(t *testing.T) {
	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.GetBcsNamespaceRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *[]bkcmdbkube.Namespace
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.GetBcsNamespaceRequest{
					CommonRequest: client.CommonRequest{
						BKBizID: 41,
						Page: client.Page{
							Limit: 100,
							Start: 0,
						},
						Fields: []string{},
						Filter: &client.PropertyFilter{
							Condition: "OR",
							Rules: []client.Rule{
								{
									Field:    "cluster_uid",
									Operator: "in",
									Value:    []string{"BCS-K8S-15171"},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			got, err := c.GetBcsNamespace(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBcsNamespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//t.Logf("GetBcsNamespace() got = %v", got)
			nsids := make([]int64, 0)
			for _, ns := range *got {
				nsids = append(nsids, ns.ID)
			}
			for _, ns := range nsids {
				fmt.Printf("%d,", ns)
			}
		})
	}
}

// Test_cmdbClient_CreateBcsNamespace tests the CreateBcsNamespace method of the cmdbClient.
func Test_cmdbClient_CreateBcsNamespace(t *testing.T) {
	bkBizID = int64(43)
	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.CreateBcsNamespaceRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *[]int
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.CreateBcsNamespaceRequest{
					BKBizID: &bkBizID,
					Data: &[]bkcmdbkube.Namespace{
						{
							ClusterSpec: bkcmdbkube.ClusterSpec{
								ClusterID: 4583,
							},
							Name: "t55est5ssssdd8ss",
							Labels: &map[string]string{
								"test": "test",
							},
							ResourceQuotas: &[]bkcmdbkube.ResourceQuota{
								{
									Hard: map[string]string{"cpu": "1"},
									ScopeSelector: &bkcmdbkube.ScopeSelector{
										MatchExpressions: []bkcmdbkube.ScopedResourceSelectorRequirement{
											{
												ScopeName: "PriorityClass",
												Operator:  "In",
												Values:    []string{"high-priority"},
											},
										},
									},
								},
							},
						},
						{
							ClusterSpec: bkcmdbkube.ClusterSpec{
								ClusterID: 4583,
							},
							Name: "test5sd55ssd8ss",
							Labels: &map[string]string{
								"test": "test",
							},
							ResourceQuotas: &[]bkcmdbkube.ResourceQuota{
								{
									Hard: map[string]string{"cpu": "1"},
									ScopeSelector: &bkcmdbkube.ScopeSelector{
										MatchExpressions: []bkcmdbkube.ScopedResourceSelectorRequirement{
											{
												ScopeName: "PriorityClass",
												Operator:  "In",
												Values:    []string{"high-priority"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			got, err := c.CreateBcsNamespace(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBcsNamespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("CreateBcsNamespace() got = %v", got)
		})
	}
}

// Test_cmdbClient_UpdateBcsNamespace tests the UpdateBcsNamespace method of the cmdbClient.
func Test_cmdbClient_UpdateBcsNamespace(t *testing.T) {
	bkBizID = int64(41)
	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.UpdateBcsNamespaceRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.UpdateBcsNamespaceRequest{
					BKBizID: &bkBizID,
					IDs:     &[]int64{56141},
					Data: &client.UpdateBcsNamespaceRequestData{
						Labels: &map[string]string{
							"testA": "testA",
						},
						ResourceQuotas: &[]bkcmdbkube.ResourceQuota{
							{
								Hard: map[string]string{"cpu": "2"},
								ScopeSelector: &bkcmdbkube.ScopeSelector{
									MatchExpressions: []bkcmdbkube.ScopedResourceSelectorRequirement{
										{
											ScopeName: "PriorityClass",
											Operator:  "In",
											Values:    []string{"high-priority"},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			if err := c.UpdateBcsNamespace(tt.args.request); (err != nil) != tt.wantErr {
				t.Errorf("UpdateBcsNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_cmdbClient_DeleteBcsNamespace tests the DeleteBcsNamespace method of the cmdbClient.
func Test_cmdbClient_DeleteBcsNamespace(t *testing.T) {
	bkBizID = int64(41)
	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.DeleteBcsNamespaceRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.DeleteBcsNamespaceRequest{
					BKBizID: &bkBizID,
					IDs:     &[]int64{4376},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			if err := c.DeleteBcsNamespace(tt.args.request); (err != nil) != tt.wantErr {
				t.Errorf("DeleteBcsNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_cmdbClient_GetBcsWorkload tests the GetBcsWorkload method of the cmdbClient.
func Test_cmdbClient_GetBcsWorkload(t *testing.T) {
	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.GetBcsWorkloadRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *[]interface{}
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.GetBcsWorkloadRequest{
					CommonRequest: client.CommonRequest{
						BKBizID: 41,
						Page: client.Page{
							Limit: 100,
							Start: 0,
						},
						Filter: &client.PropertyFilter{
							Condition: "OR",
							Rules: []client.Rule{
								{
									Field:    "cluster_uid",
									Operator: "in",
									Value:    []string{"BCS-K8S-15276"},
								},
							},
						},
					},
					//ClusterUID: "cluster-bcs",
					Kind: "deployment",
					//BKClusterID: 449,
					//Namespace:  "test",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.GetBcsWorkloadRequest{
					CommonRequest: client.CommonRequest{
						BKBizID: 41,
						Page: client.Page{
							Limit: 100,
							Start: 0,
						},
						Filter: &client.PropertyFilter{
							Condition: "OR",
							Rules: []client.Rule{
								{
									Field:    "bk_cluster_id",
									Operator: "in",
									Value:    []int64{879},
								},
							},
						},
					},
					//ClusterUID: "cluster-bcs",
					Kind: "daemonSet",
					//BKClusterID: 449,
					//Namespace:  "test",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.GetBcsWorkloadRequest{
					CommonRequest: client.CommonRequest{
						BKBizID: 41,
						Page: client.Page{
							Limit: 100,
							Start: 0,
						},
						Filter: &client.PropertyFilter{
							Condition: "OR",
							Rules: []client.Rule{
								{
									Field:    "bk_cluster_id",
									Operator: "in",
									Value:    []int64{879},
								},
							},
						},
					},
					//ClusterUID: "cluster-bcs",
					Kind: "statefulSet",
					//BKClusterID: 449,
					//Namespace:  "test",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.GetBcsWorkloadRequest{
					CommonRequest: client.CommonRequest{
						BKBizID: 41,
						Page: client.Page{
							Limit: 100,
							Start: 0,
						},
						Filter: &client.PropertyFilter{
							Condition: "OR",
							Rules: []client.Rule{
								{
									Field:    "bk_cluster_id",
									Operator: "in",
									Value:    []int64{879},
								},
							},
						},
					},
					//ClusterUID: "cluster-bcs",
					Kind: "gameDeployment",
					//BKClusterID: 449,
					//Namespace:  "test",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.GetBcsWorkloadRequest{
					CommonRequest: client.CommonRequest{
						BKBizID: 41,
						Page: client.Page{
							Limit: 100,
							Start: 0,
						},
						Filter: &client.PropertyFilter{
							Condition: "OR",
							Rules: []client.Rule{
								{
									Field:    "bk_cluster_id",
									Operator: "in",
									Value:    []int64{879},
								},
							},
						},
					},
					//ClusterUID: "cluster-bcs",
					Kind: "gameStatefulSet",
					//BKClusterID: 449,
					//Namespace:  "test",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.GetBcsWorkloadRequest{
					CommonRequest: client.CommonRequest{
						BKBizID: 41,
						Page: client.Page{
							Limit: 100,
							Start: 0,
						},
						Filter: &client.PropertyFilter{
							Condition: "OR",
							Rules: []client.Rule{
								{
									Field:    "id",
									Operator: "in",
									Value:    []int64{10},
								},
							},
						},
					},
					//ClusterUID: "cluster-bcs",
					Kind: "daemonSet",
					//BKClusterID: 449,
					//Namespace:  "test",
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			got, err := c.GetBcsWorkload(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBcsWorkload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("GetBcsWorkload() got = %v", got)

			workloadids := make([]int64, 0)
			for _, v := range *got {
				workloadids = append(workloadids, (int64)(v.(map[string]interface{})["id"].(float64)))
			}

			for _, v := range workloadids {
				fmt.Printf("%d,", v)
			}

		})
	}
}

// Test_cmdbClient_CreateBcsWorkload tests the CreateBcsWorkload method of the cmdbClient.
func Test_cmdbClient_CreateBcsWorkload(t *testing.T) {
	bkBizID = int64(43)
	kind := "deployment"
	nsid := int64(56141)
	name := "deployment1"
	replicas := int64(0)
	minReadySeconds := int64(0)
	strategyType := ""

	rud := bkcmdbkube.RollingUpdateDeployment{
		MaxUnavailable: &bkcmdbkube.IntOrString{
			Type:   0,
			IntVal: 1,
			StrVal: "123",
		},
		MaxSurge: &bkcmdbkube.IntOrString{
			Type:   0,
			IntVal: 1,
			StrVal: "123",
		},
	}

	jsonBytes, err := json.Marshal(rud)
	if err != nil {
		t.Errorf("CreateBcsWorkload() error = %v", err)
		return
	}

	rudMap := make(map[string]interface{})
	_ = json.Unmarshal(jsonBytes, &rudMap)

	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.CreateBcsWorkloadRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *[]int64
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.CreateBcsWorkloadRequest{
					BKBizID: &bkBizID,
					Kind:    &kind,
					Data: &[]client.CreateBcsWorkloadRequestData{
						{
							NamespaceID: &nsid,
							Name:        &name,
							Labels: &map[string]string{
								"app": "test",
							},
							Selector: &bkcmdbkube.LabelSelector{
								MatchLabels: map[string]string{
									"app": "test",
								},
								MatchExpressions: []bkcmdbkube.LabelSelectorRequirement{
									{
										Key:      "app",
										Operator: "In",
										Values:   []string{"test"},
									},
								},
							},
							Replicas:              &replicas,
							MinReadySeconds:       &minReadySeconds,
							StrategyType:          &strategyType,
							RollingUpdateStrategy: &rudMap,
						},
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			got, err := c.CreateBcsWorkload(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBcsWorkload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("CreateBcsWorkload() got = %v", got)
		})
	}
}

// Test_cmdbClient_UpdateBcsWorkload tests the UpdateBcsWorkload method of the cmdbClient.
func Test_cmdbClient_UpdateBcsWorkload(t *testing.T) {
	kind := "deployment"
	replicas := int64(0)
	minReadySeconds := int64(0)
	strategyType := "Always"

	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.UpdateBcsWorkloadRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.UpdateBcsWorkloadRequest{
					BKBizID: &bkBizID,
					Kind:    &kind,
					IDs:     &[]int64{246030},
					Data: &client.UpdateBcsWorkloadRequestData{
						Labels: &map[string]string{
							"app": "testaaaa",
						},
						Selector: &bkcmdbkube.LabelSelector{
							MatchLabels: map[string]string{
								"app": "testaaaa",
							},
						},
						Replicas:              &replicas,
						MinReadySeconds:       &minReadySeconds,
						StrategyType:          &strategyType,
						RollingUpdateStrategy: nil,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			if err := c.UpdateBcsWorkload(tt.args.request); (err != nil) != tt.wantErr {
				t.Errorf("UpdateBcsWorkload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_cmdbClient_DeleteBcsWorkload tests the DeleteBcsWorkload method of the cmdbClient.
func Test_cmdbClient_DeleteBcsWorkload(t *testing.T) {
	kind := "deployment"
	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.DeleteBcsWorkloadRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.DeleteBcsWorkloadRequest{
					BKBizID: &bkBizID,
					Kind:    &kind,
					IDs: &[]int64{
						246030,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			if err := c.DeleteBcsWorkload(tt.args.request); (err != nil) != tt.wantErr {
				t.Errorf("DeleteBcsWorkload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_cmdbClient_GetBcsNode tests the GetBcsNode method of the cmdbClient.
func Test_cmdbClient_GetBcsNode(t *testing.T) {
	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.GetBcsNodeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *[]bkcmdbkube.Node
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.GetBcsNodeRequest{
					CommonRequest: client.CommonRequest{
						BKBizID: 41,
						Page: client.Page{
							Limit: 100,
							Start: 0,
						},
						Filter: &client.PropertyFilter{
							Condition: "OR",
							Rules: []client.Rule{
								{
									Field:    "bk_cluster_id",
									Operator: "in",
									Value:    []int64{4625},
								},
							},
						},
					},
					//BKClusterID: 449,
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			got, err := c.GetBcsNode(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBcsNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//t.Logf("GetBcsNode() got = %v", got)

			nodeids := make([]int64, 0)
			for _, node := range *got {
				nodeids = append(nodeids, node.ID)
			}
			for _, nodeid := range nodeids {
				fmt.Printf("%d,", nodeid)
			}
		})
	}
}

// Test_cmdbClient_CreateBcsNode tests the CreateBcsNode method of the cmdbClient.
func Test_cmdbClient_CreateBcsNode(t *testing.T) {
	hostID := int64(1250)
	clusterID := int64(4583)
	name := "12tesssssss0ssssst"
	unschedulable := false
	hostName := ""
	runtimeComponent := ""
	kubeProxyMode := ""
	podCidr := ""

	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.CreateBcsNodeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *[]int64
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.CreateBcsNodeRequest{
					BKBizID: &bkBizID,
					Data: &[]client.CreateBcsNodeRequestData{
						{
							HostID:    &hostID,
							ClusterID: &clusterID,
							Name:      &name,
							Labels: &map[string]string{
								"test": "test",
							},
							Taints: &map[string]string{
								"test": "test",
							},
							Unschedulable: &unschedulable,
							InternalIP: &[]string{
								"1.1.1.1",
							},
							ExternalIP: &[]string{
								"1.1.1.1",
							},
							HostName:         &hostName,
							RuntimeComponent: &runtimeComponent,
							KubeProxyMode:    &kubeProxyMode,
							PodCidr:          &podCidr,
						},
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			got, err := c.CreateBcsNode(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBcsNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("CreateBcsNode() got = %v", got)
		})
	}
}

// Test_cmdbClient_UpdateBcsNode tests the UpdateBcsNode method of the cmdbClient.
func Test_cmdbClient_UpdateBcsNode(t *testing.T) {
	unschedulable := false

	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.UpdateBcsNodeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.UpdateBcsNodeRequest{
					BKBizID: &bkBizID,
					IDs: &[]int64{
						655054,
					},
					Data: &client.UpdateBcsNodeRequestData{
						Labels: &map[string]string{
							"test": "test",
						},
						Taints: &map[string]string{
							"test": "test",
						},
						Unschedulable: &unschedulable,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			if err := c.UpdateBcsNode(tt.args.request); (err != nil) != tt.wantErr {
				t.Errorf("UpdateBcsNode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_cmdbClient_DeleteBcsNode tests the DeleteBcsNode method of the cmdbClient.
func Test_cmdbClient_DeleteBcsNode(t *testing.T) {
	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.DeleteBcsNodeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.DeleteBcsNodeRequest{
					BKBizID: &bkBizID,
					IDs: &[]int64{
						614991,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			if err := c.DeleteBcsNode(tt.args.request); (err != nil) != tt.wantErr {
				t.Errorf("DeleteBcsNode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_cmdbClient_GetBcsPod tests the GetBcsPod method of the cmdbClient.
func Test_cmdbClient_GetBcsPod(t *testing.T) {
	bkBizID = int64(41)
	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.GetBcsPodRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *[]bkcmdbkube.Pod
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.GetBcsPodRequest{
					CommonRequest: client.CommonRequest{
						BKBizID: bkBizID,
						Page: client.Page{
							Limit: 200,
							Start: 0,
						},
						Filter: &client.PropertyFilter{
							Condition: "AND",
							Rules: []client.Rule{
								{
									Field:    "cluster_uid",
									Operator: "in",
									Value:    []string{"BCS-K8S-910215519833"},
								},
							},
						},
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			got, err := c.GetBcsPod(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBcsPod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("GetBcsPod() got = %v", got)

			podids := make([]int64, 0)
			for _, pod := range *got {
				podids = append(podids, pod.ID)
			}
			for _, podid := range podids {
				fmt.Printf("%d,", podid)
			}
		})
	}
}

// Test_cmdbClient_CreateBcsPod tests the CreateBcsPod method of the cmdbClient.
func Test_cmdbClient_CreateBcsPod(t *testing.T) {
	bkBizID = int64(41)
	clusterID := int64(4625)
	namespaceID := int64(56331)
	workloadKind := "deployment"
	workloadID := int64(246735)
	nodeID := int64(655262)
	name := "event-exporter-dddc48bf9-9mns4sss2"
	hostID := int64(1250)
	priority := int32(0)
	ip := "10.0.0.1"
	operator := []string{""}

	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.CreateBcsPodRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *[]int64
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.CreateBcsPodRequest{
					Data: &[]client.CreateBcsPodRequestData{
						{
							BizID: &bkBizID,
							Pods: &[]client.CreateBcsPodRequestDataPod{
								{
									Spec: &client.CreateBcsPodRequestPodSpec{
										ClusterID:    &clusterID,
										NameSpaceID:  &namespaceID,
										WorkloadKind: &workloadKind,
										WorkloadID:   &workloadID,
										NodeID:       &nodeID,
										Ref: &bkcmdbkube.Reference{
											Kind: "deployment",
											Name: "test-deploy",
											ID:   246735,
										},
									},

									Name:     &name,
									HostID:   &hostID,
									Priority: &priority,
									Operator: &operator,
									Labels: &map[string]string{
										"test": "test",
									},
									IP: &ip,
									IPs: &[]bkcmdbkube.PodIP{
										{
											IP: "1,1,1,1",
										},
									},
									Containers: &[]bkcmdbkube.ContainerBaseFields{},
								},
							},
						},
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			got, err := c.CreateBcsPod(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBcsPod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("CreateBcsPod() got = %v", got)
		})
	}
}

// Test_cmdbClient_DeleteBcsPod tests the DeleteBcsPod method of the cmdbClient.
func Test_cmdbClient_DeleteBcsPod(t *testing.T) {
	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.DeleteBcsPodRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{},
			args: args{
				request: &client.DeleteBcsPodRequest{
					Data: &[]client.DeleteBcsPodRequestData{
						{
							BKBizID: &bkBizID,
							IDs: &[]int64{
								2560,
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			if err := c.DeleteBcsPod(tt.args.request); (err != nil) != tt.wantErr {
				t.Errorf("DeleteBcsPod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_deleteAllByBkBizID tests delete all bcs resources by bizid
func Test_deleteAllByBkBizID(t *testing.T) {
	bkBizID = int64(43)
	c := getCli()
	t.Logf("start delete all")
	t.Logf("start delete all pod")
	for {
		got, err := c.GetBcsPod(&client.GetBcsPodRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bkBizID,
				Fields:  []string{"id"},
				Page: client.Page{
					Limit: 200,
					Start: 0,
				},
			},
		})
		if err != nil {
			t.Errorf("GetBcsPod() error = %v", err)
			return
		}
		podToDelete := make([]int64, 0)
		for _, pod := range *got {
			podToDelete = append(podToDelete, pod.ID)
		}

		if len(podToDelete) == 0 {
			break
		} else {
			t.Logf("delete pod: %v", podToDelete)
			err := c.DeleteBcsPod(&client.DeleteBcsPodRequest{
				Data: &[]client.DeleteBcsPodRequestData{
					{
						BKBizID: &bkBizID,
						IDs:     &podToDelete,
					},
				},
			})
			if err != nil {
				t.Errorf("DeleteBcsPod() error = %v", err)
				return
			}
		}
	}
	t.Logf("delete all pod success")

	t.Logf("start delete all workload")
	workloadTypes := []string{"deployment", "statefulSet", "daemonSet", "gameDeployment", "gameStatefulSet", "pods"}

	for _, workloadType := range workloadTypes {
		for {
			got, err := c.GetBcsWorkload(&client.GetBcsWorkloadRequest{
				CommonRequest: client.CommonRequest{
					BKBizID: bkBizID,
					Fields:  []string{"id"},
					Page: client.Page{
						Limit: 200,
						Start: 0,
					},
				},
				Kind: workloadType,
			})
			if err != nil {
				t.Errorf("GetBcsWorkload() error = %v", err)
				return
			}
			workloadToDelete := make([]int64, 0)
			for _, workload := range *got {
				workloadToDelete = append(workloadToDelete, (int64)(workload.(map[string]interface{})["id"].(float64)))
			}

			if len(workloadToDelete) == 0 {
				break
			} else {
				t.Logf("delete workload: %v", workloadToDelete)
				err := c.DeleteBcsWorkload(&client.DeleteBcsWorkloadRequest{
					BKBizID: &bkBizID,
					Kind:    &workloadType,
					IDs:     &workloadToDelete,
				})
				if err != nil {
					t.Errorf("DeleteBcsWorkload() error = %v", err)
					return
				}
			}
		}
	}
	t.Logf("delete all workload success")

	t.Logf("start delete all namespace")
	for {
		got, err := c.GetBcsNamespace(&client.GetBcsNamespaceRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bkBizID,
				Fields:  []string{"id"},
				Page: client.Page{
					Limit: 200,
					Start: 0,
				},
			},
		})
		if err != nil {
			t.Errorf("GetBcsNamespace() error = %v", err)
			return
		}
		namespaceToDelete := make([]int64, 0)
		for _, namespace := range *got {
			namespaceToDelete = append(namespaceToDelete, namespace.ID)
		}

		if len(namespaceToDelete) == 0 {
			break
		} else {
			t.Logf("delete namespace: %v", namespaceToDelete)
			err := c.DeleteBcsNamespace(&client.DeleteBcsNamespaceRequest{
				BKBizID: &bkBizID,
				IDs:     &namespaceToDelete,
			})
			if err != nil {
				t.Errorf("DeleteBcsNamespace() error = %v", err)
				return
			}
		}
	}
	t.Logf("delete all namespace success")

	t.Logf("start delete all node")
	for {
		got, err := c.GetBcsNode(&client.GetBcsNodeRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bkBizID,
				Page: client.Page{
					Limit: 100,
					Start: 0,
				},
			},
		})
		if err != nil {
			t.Errorf("GetBcsNode() error = %v", err)
			return
		}
		nodeToDelete := make([]int64, 0)
		for _, node := range *got {
			nodeToDelete = append(nodeToDelete, node.ID)
		}

		if len(nodeToDelete) == 0 {
			break
		} else {
			t.Logf("delete node: %v", nodeToDelete)
			err := c.DeleteBcsNode(&client.DeleteBcsNodeRequest{
				BKBizID: &bkBizID,
				IDs:     &nodeToDelete,
			})
			if err != nil {
				t.Errorf("DeleteBcsNode() error = %v", err)
				return
			}
		}
	}
	t.Logf("delete all node success")

	t.Logf("start delete all cluster")
	for {
		got, err := c.GetBcsCluster(&client.GetBcsClusterRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bkBizID,
				Fields:  []string{"id"},
				Page: client.Page{
					Limit: 10,
					Start: 0,
				},
			},
		})
		if err != nil {
			t.Errorf("GetBcsCluster() error = %v", err)
			return
		}
		clusterToDelete := make([]int64, 0)
		for _, cluster := range *got {
			clusterToDelete = append(clusterToDelete, cluster.ID)
		}

		if len(clusterToDelete) == 0 {
			break
		} else {
			t.Logf("delete cluster: %v", clusterToDelete)
			err := c.DeleteBcsCluster(&client.DeleteBcsClusterRequest{
				BKBizID: &bkBizID,
				IDs:     &clusterToDelete,
			})
			if err != nil {
				t.Errorf("DeleteBcsCluster() error = %v", err)
				return
			}
		}
	}
	t.Logf("delete all cluster success")
	t.Logf("delete all success")
}

// Test_getAllByBkBizID tests get all bcs resources by bizid
func Test_getAllByBkBizID(t *testing.T) {
	workloadTypes := []string{"deployment", "statefulSet", "daemonSet", "gameDeployment", "gameStatefulSet"}
	c := getCli()
	t.Logf("start get all cluster")
	clusters := make(map[int64]string, 0)
	clusterPage := 0
	for {
		clusterGot, err := c.GetBcsCluster(&client.GetBcsClusterRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bkBizID,
				Page: client.Page{
					Limit: 100 * (clusterPage + 1),
					Start: 100 * clusterPage,
				},
			},
		})
		if err != nil {
			t.Errorf("GetBcsCluster() error = %v", err)
			return
		}

		for _, cluster := range *clusterGot {
			clusters[cluster.ID] = cluster.Uid
		}

		if len(*clusterGot) < 100 {
			break
		}
		clusterPage++
	}

	if len(clusters) == 0 {
		t.Logf("no cluster found")
		return
	}

	t.Logf("get cluster: %v", clusters)

	for clusterID, clusterUID := range clusters {
		t.Logf("=======================================")
		t.Logf("clusterID: %d, clusterUID: %s", clusterID, clusterUID)
		nodes := make(map[int64]string, 0)
		nodePage := 0
		for {
			nodeGot, err := c.GetBcsNode(&client.GetBcsNodeRequest{
				CommonRequest: client.CommonRequest{
					BKBizID: bkBizID,
					Page: client.Page{
						Limit: 100 * (nodePage + 1),
						Start: 100 * nodePage,
					},
					Filter: &client.PropertyFilter{
						Condition: "OR",
						Rules: []client.Rule{
							{
								Field:    "bk_cluster_id",
								Operator: "in",
								Value:    []int64{clusterID},
							},
						},
					},
				},
				//BKClusterID: clusterID,
			})
			if err != nil {
				t.Errorf("GetBcsNode() error = %v", err)
				return
			}

			for _, node := range *nodeGot {
				nodes[node.ID] = *node.Name
			}

			if len(*nodeGot) < 100 {
				break
			}
			nodePage++
		}

		if len(nodes) == 0 {
			t.Logf("clusterID: %d, clusterUID: %s, nodes: no node found", clusterID, clusterUID)
		} else {
			t.Logf("clusterID: %d, clusterUID: %s, nodes: %v", clusterID, clusterUID, nodes)
		}

		namespaces := make(map[int64]string, 0)
		namespacePage := 0
		for {
			namespaceGot, err := c.GetBcsNamespace(&client.GetBcsNamespaceRequest{
				CommonRequest: client.CommonRequest{
					BKBizID: bkBizID,
					Page: client.Page{
						Limit: 100 * (namespacePage + 1),
						Start: 100 * namespacePage,
					},
					Filter: &client.PropertyFilter{
						Condition: "OR",
						Rules: []client.Rule{
							{
								Field:    "cluster_uid",
								Operator: "in",
								Value:    []string{clusterUID},
							},
						},
					},
				},
				//ClusterUID: clusterUID,
			})
			if err != nil {
				t.Errorf("GetBcsNamespace() error = %v", err)
				return
			}

			for _, namespace := range *namespaceGot {
				namespaces[namespace.ID] = namespace.Name
			}

			if len(*namespaceGot) < 100 {
				break
			}
			namespacePage++
		}

		if len(namespaces) == 0 {
			t.Logf("clusterID: %d, clusterUID: %s, namespaces: no namespace found", clusterID, clusterUID)
		} else {
			t.Logf("clusterID: %d, clusterUID: %s, namespaces: %v", clusterID, clusterUID, namespaces)
		}

		for _, namespaceName := range namespaces {
			for _, workloadType := range workloadTypes {
				workloads := make(map[int64]string, 0)
				workloadPage := 0
				for {
					workloadGot, err := c.GetBcsWorkload(&client.GetBcsWorkloadRequest{
						CommonRequest: client.CommonRequest{
							BKBizID: bkBizID,
							Page: client.Page{
								Limit: 100 * (workloadPage + 1),
								Start: 100 * workloadPage,
							},
							Filter: &client.PropertyFilter{
								Condition: "AND",
								Rules: []client.Rule{
									{
										Field:    "cluster_uid",
										Operator: "in",
										Value:    []string{clusterUID},
									},
									{
										Field:    "namespace",
										Operator: "in",
										Value:    []string{namespaceName},
									},
								},
							},
						},

						//ClusterUID: clusterUID,
						//Namespace:  namespaceName,
						Kind: workloadType,
					})
					if err != nil {
						t.Errorf("GetBcsWorkload() error = %v", err)
						return
					}

					for _, workload := range *workloadGot {
						workloads[(int64)(workload.(map[string]interface{})["id"].(float64))] = workload.(map[string]interface{})["name"].(string)
					}

					if len(*workloadGot) < 100 {
						break
					}
					workloadPage++
				}

				if len(workloads) == 0 {
					t.Logf("clusterID: %d, clusterUID: %s, namespace: %s, workloadType: %s, workloads: no workload found", clusterID, clusterUID, namespaceName, workloadType)
				} else {
					t.Logf("clusterID: %d, clusterUID: %s, namespace: %s, workloadType: %s, workloads: %v", clusterID, clusterUID, namespaceName, workloadType, workloads)
				}

				for workloadID, workloadName := range workloads {
					pods := make(map[int64]string, 0)
					podPage := 0
					for {
						podGot, err := c.GetBcsPod(&client.GetBcsPodRequest{
							CommonRequest: client.CommonRequest{
								BKBizID: bkBizID,
								Page: client.Page{
									Limit: 100 * (podPage + 1),
									Start: 100 * podPage,
								},
								Filter: &client.PropertyFilter{
									Condition: "AND",
									Rules: []client.Rule{
										{
											Field:    "ref.id",
											Operator: "in",
											Value:    []int64{workloadID},
										},
										{
											Field:    "cluster_uid",
											Operator: "in",
											Value:    []string{clusterUID},
										},
										{
											Field:    "namespace",
											Operator: "in",
											Value:    []string{namespaceName},
										},
									},
								},
							},
							//ClusterUID: clusterUID,
							//Namespace:  namespaceName,
						})
						if err != nil {
							t.Errorf("GetBcsPod() error = %v", err)
							return
						}

						for _, pod := range *podGot {
							pods[pod.ID] = *pod.Name
						}

						if len(*podGot) < 100 {
							break
						}
						podPage++
					}

					if len(pods) == 0 {
						t.Logf("clusterID: %d, clusterUID: %s, namespace: %s, workloadType: %s, workloadID: %d, workloadName: %s, pods: no pod found", clusterID, clusterUID, namespaceName, workloadType, workloadID, workloadName)
					} else {
						t.Logf("clusterID: %d, clusterUID: %s, namespace: %s, workloadType: %s, workloadID: %d, workloadName: %s, pods: %v", clusterID, clusterUID, namespaceName, workloadType, workloadID, workloadName, pods)
					}
				}
			}
		}

		t.Logf("=======================================")
	}
}

// Test_cmdbClient_UpdateBcsClusterType tests the UpdateBcsClusterType method of the cmdbClient.
// nolint
func Test_cmdbClient_UpdateBcsClusterType(t *testing.T) {
	id := int64(4583)
	clusterType := "SHARE_CLUSTER"
	type fields struct {
		config   *Options
		userAuth string
	}
	type args struct {
		request *client.UpdateBcsClusterTypeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				request: &client.UpdateBcsClusterTypeRequest{
					BKBizID: &bkBizID,
					ID:      &id,
					Type:    &clusterType,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getCli()
			if err := c.UpdateBcsClusterType(tt.args.request); (err != nil) != tt.wantErr {
				t.Errorf("UpdateBcsClusterType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
