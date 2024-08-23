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

package permitcheck

import (
	"context"
	"net/http"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"
)

// CheckClusterPermission check cluster permission
func (c *checker) CheckClusterPermission(ctx context.Context, query *cluster.ClusterQuery, action RSAction) (
	*v1alpha1.Cluster, int, error) {
	argoCluster, err := c.store.GetCluster(ctx, query)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "get cluster from storage failure")
	}
	if argoCluster == nil {
		return nil, http.StatusBadRequest, errors.Errorf("cluster '%v' not found", query)
	}

	var statusCode int
	_, statusCode, err = c.CheckProjectPermission(ctx, argoCluster.Project, ProjectViewRSAction)
	if err != nil {
		return nil, statusCode, err
	}
	return argoCluster, http.StatusOK, nil
}

func (c *checker) getMultiClustersMultiActionsPermission(ctx context.Context, project string, clusters []string) (
	[]interface{}, *UserResourcePermission, int, error) {
	resultClusters := make([]interface{}, 0, len(clusters))
	projClusters := make(map[string][]*v1alpha1.Cluster)
	// list all clusters by project
	if project != "" && len(clusters) == 0 {
		argoClusterList, err := c.store.ListClustersByProjectName(ctx, project)
		if err != nil {
			return nil, nil, http.StatusInternalServerError, err
		}
		for i := range argoClusterList.Items {
			argoCluster := argoClusterList.Items[i]
			resultClusters = append(resultClusters, &argoCluster)
			projClusters[project] = append(projClusters[project], &argoCluster)
		}
	} else {
		for i := range clusters {
			cls := clusters[i]
			argoCluster, err := c.store.GetCluster(ctx, &cluster.ClusterQuery{
				Name: cls,
			})
			if err != nil {
				return nil, nil, http.StatusInternalServerError, errors.Wrapf(err, "get cluster failed")
			}
			if argoCluster == nil {
				return nil, nil, http.StatusBadRequest, errors.Errorf("cluster '%s' not found", cls)
			}
			proj := argoCluster.Project
			_, ok := projClusters[proj]
			if ok {
				projClusters[proj] = append(projClusters[proj], argoCluster)
			} else {
				projClusters[proj] = []*v1alpha1.Cluster{argoCluster}
			}
			resultClusters = append(resultClusters, argoCluster)
		}
	}

	result := &UserResourcePermission{
		ResourceType:  ClusterRSType,
		ResourcePerms: make(map[string]map[RSAction]bool),
		ActionPerms:   map[RSAction]bool{ClusterViewRSAction: true},
	}
	if len(resultClusters) == 0 {
		return resultClusters, result, http.StatusOK, nil
	}

	allView := false
	for proj, argoClusters := range projClusters {
		_, projPermits, statusCode, err := c.getProjectMultiActionsPermission(ctx, proj)
		if statusCode != http.StatusForbidden && statusCode != http.StatusOK {
			return nil, nil, statusCode, err
		}
		if projPermits[ProjectViewRSAction] {
			allView = true
		}
		for _, cls := range argoClusters {
			result.ResourcePerms[cls.Name] = map[RSAction]bool{
				ClusterViewRSAction: projPermits[ProjectViewRSAction],
			}
		}
	}
	result.ActionPerms = map[RSAction]bool{
		ClusterViewRSAction: allView,
	}
	return resultClusters, result, http.StatusOK, nil
}
