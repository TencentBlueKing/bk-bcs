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

// Package cmdb cmdb operate
package cmdb

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// CmdbAction cmdb action interface
type CmdbAction interface { // nolint
	DeleteAllByBkBizIDAndBkClusterID(ctx context.Context, req *types.DeleteAllByBkBizIDAndBkClusterIDReq) (bool, error)
	GetBusiness(ctx context.Context) (*[]cmdb.GetBusinessRespDataInfo, error)
}

// Action action for cmdb
type Action struct{}

// NewCmdbAction new cmdb action
func NewCmdbAction() CmdbAction {
	return &Action{}
}

// DeleteAllByBkBizIDAndBkClusterID delete all by bk_biz_id and bk_cluster_id
func (a *Action) DeleteAllByBkBizIDAndBkClusterID(ctx context.Context,
	req *types.DeleteAllByBkBizIDAndBkClusterIDReq) (bool, error) {
	cmdbClient := cmdb.GetCmdbClient()

	result := false

	err := deleteAll2IDPod(req.BkBizID, req.BkClusterID, cmdbClient)
	if err != nil {
		return result, err
	}

	err = deleteAll2IDWorkload(req.BkBizID, req.BkClusterID, cmdbClient)
	if err != nil {
		return result, err
	}

	deleteAll2IDNamespace(req.BkBizID, req.BkClusterID, cmdbClient)
	deleteAll2IDIDNode(req.BkBizID, req.BkClusterID, cmdbClient)
	deleteAll2IDCluster(req.BkBizID, req.BkClusterID, cmdbClient)

	result = true
	return result, nil
}

// GetBusiness get business
func (a *Action) GetBusiness(ctx context.Context) (*[]cmdb.GetBusinessRespDataInfo, error) {
	data, err := cmdb.GetCmdbClient().GetBusiness()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func deleteAll2IDPod(bkBizID int64, bkClusterID []int64, c *cmdb.Client) error {
	blog.Info("start delete all pod")

	for {
		got, err := c.GetBcsPod(&cmdb.GetBcsPodReq{
			CommonReq: cmdb.CommonReq{
				BKBizID: bkBizID,
				Fields:  []string{"id"},
				Page: cmdb.Page{
					Start: 0,
					Limit: 200,
				},
				Filter: &cmdb.PropertyFilter{
					Condition: "AND",
					Rules: []cmdb.Rule{
						{
							Field:    "bk_cluster_id",
							Operator: "in",
							Value:    bkClusterID,
						},
					},
				},
			},
		})
		if err != nil {
			return err
		}

		podToDelete := make([]int64, 0)
		for _, pod := range *got {
			podToDelete = append(podToDelete, pod.ID)
		}

		if len(podToDelete) == 0 {
			break
		}

		blog.Info("delete pod %v", podToDelete)
		err = c.DeleteBcsPod(&cmdb.DeleteBcsPodReq{
			Data: &[]cmdb.DeleteBcsPodReqData{
				{
					BKBizID: &bkBizID,
					IDs:     &podToDelete,
				},
			},
		})
		if err != nil {
			blog.Errorf("DeleteBcsPod err: %v", err)
			return err
		}
	}

	blog.Info("delete all pod success")

	return nil
}

func deleteAll2IDWorkload(bkBizID int64, bkClusterID []int64, c *cmdb.Client) error {
	blog.Info("start delete all workload")

	workloadTypes := []string{"deployment", "statefulSet", "daemonSet", "gameDeployment", "gameStatefulSet", "pods"}
	for _, workloadType := range workloadTypes {
		for {
			got, err := c.GetBcsWorkload(&cmdb.GetBcsWorkloadReq{
				CommonReq: cmdb.CommonReq{
					BKBizID: bkBizID,
					Fields:  []string{"id"},
					Page: cmdb.Page{
						Start: 0,
						Limit: 200,
					},
					Filter: &cmdb.PropertyFilter{
						Condition: "AND",
						Rules: []cmdb.Rule{
							{
								Field:    "bk_cluster_id",
								Operator: "in",
								Value:    bkClusterID,
							},
						},
					},
				},
				Kind: workloadType,
			})
			if err != nil {
				blog.Errorf("GetBcsWorkload err: %v", err)
				return err
			}

			workloadToDelete := make([]int64, 0)
			for _, workload := range *got {
				workloadToDelete = append(workloadToDelete, (int64)(workload.(map[string]interface{})["id"].(float64)))
			}

			if len(workloadToDelete) == 0 {
				break
			}

			blog.Infof("delete workload: %v", workloadToDelete)
			err = c.DeleteBcsWorkload(&cmdb.DeleteBcsWorkloadReq{
				BKBizID: &bkBizID,
				Kind:    &workloadType,
				IDs:     &workloadToDelete,
			})
			if err != nil {
				blog.Errorf("DeleteBcsWorkload err: %v", err)
				return err
			}
		}
	}

	blog.Info("delete all workload success")

	return nil
}

func deleteAll2IDNamespace(bkBizID int64, bkClusterID []int64, c *cmdb.Client) {
	blog.Info("start delete all namespace")

	for {
		got, err := c.GetBcsNamespace(&cmdb.GetBcsNamespaceReq{
			CommonReq: cmdb.CommonReq{
				BKBizID: bkBizID,
				Fields:  []string{"id"},
				Page: cmdb.Page{
					Limit: 200,
					Start: 0,
				},
				Filter: &cmdb.PropertyFilter{
					Condition: "AND",
					Rules: []cmdb.Rule{
						{
							Field:    "bk_cluster_id",
							Operator: "in",
							Value:    bkClusterID,
						},
					},
				},
			},
		})
		if err != nil {
			blog.Errorf("GetBcsNamespace err: %v", err)
			return
		}

		namespaceToDelete := make([]int64, 0)
		for _, namespace := range *got {
			namespaceToDelete = append(namespaceToDelete, namespace.ID)
		}

		if len(namespaceToDelete) == 0 {
			break
		}

		blog.Infof("delete namespace: %v", namespaceToDelete)
		err = c.DeleteBcsNamespace(&cmdb.DeleteBcsNamespaceReq{
			BKBizID: &bkBizID,
			IDs:     &namespaceToDelete,
		})
		if err != nil {
			blog.Errorf("DeleteBcsNamespace() error = %v", err)
			return
		}
	}

	blog.Info("delete all namespace success")
}

func deleteAll2IDIDNode(bkBizID int64, bkClusterID []int64, c *cmdb.Client) {
	blog.Info("start delete all node")

	for {
		got, err := c.GetBcsNode(&cmdb.GetBcsNodeReq{
			CommonReq: cmdb.CommonReq{
				BKBizID: bkBizID,
				Page: cmdb.Page{
					Limit: 100,
					Start: 0,
				},
				Filter: &cmdb.PropertyFilter{
					Condition: "AND",
					Rules: []cmdb.Rule{
						{
							Field:    "bk_cluster_id",
							Operator: "in",
							Value:    bkClusterID,
						},
					},
				},
			},
		})
		if err != nil {
			blog.Errorf("GetBcsNode err: %v", err)
			return
		}
		nodeToDelete := make([]int64, 0)
		for _, node := range *got {
			nodeToDelete = append(nodeToDelete, node.ID)
		}

		if len(nodeToDelete) == 0 {
			break
		}

		blog.Infof("delete node: %v", nodeToDelete)
		err = c.DeleteBcsNode(&cmdb.DeleteBcsNodeReq{
			BKBizID: &bkBizID,
			IDs:     &nodeToDelete,
		})
		if err != nil {
			blog.Errorf("DeleteBcsNode err: %v", err)
			return
		}
	}

	blog.Info("delete all node success")
}

func deleteAll2IDCluster(bkBizID int64, bkClusterID []int64, c *cmdb.Client) {
	blog.Info("start delete all cluster")

	for {
		got, err := c.GetBcsCluster(&cmdb.GetBcsClusterReq{
			CommonReq: cmdb.CommonReq{
				BKBizID: bkBizID,
				Fields:  []string{"id"},
				Page: cmdb.Page{
					Limit: 10,
					Start: 0,
				},
				Filter: &cmdb.PropertyFilter{
					Condition: "AND",
					Rules: []cmdb.Rule{
						{
							Field:    "id",
							Operator: "in",
							Value:    bkClusterID,
						},
					},
				},
			},
		})
		if err != nil {
			blog.Errorf("GetBcsCluster err: %v", err)
			return
		}
		clusterToDelete := make([]int64, 0)
		for _, cluster := range *got {
			clusterToDelete = append(clusterToDelete, cluster.ID)
		}

		if len(clusterToDelete) == 0 {
			break
		}

		blog.Infof("delete cluster: %v", clusterToDelete)
		err = c.DeleteBcsCluster(&cmdb.DeleteBcsClusterReq{
			BKBizID: &bkBizID,
			IDs:     &clusterToDelete,
		})
		if err != nil {
			blog.Errorf("DeleteBcsCluster err: %v", err)
			return
		}
	}

	blog.Info("delete all cluster success")
}
