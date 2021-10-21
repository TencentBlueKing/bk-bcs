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

package sqlstore

import (
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
	"strings"

	"github.com/dchest/uniuri"
)

// GetCluster query for the cluster by given clusterId
func GetCluster(clusterId string) *m.Cluster {
	cluster := m.Cluster{}
	GCoreDB.Where(&m.Cluster{ID: clusterId}).First(&cluster)
	if cluster.ID != "" {
		return &cluster
	}
	return nil
}

// GetAllCluster query for all clusters
func GetAllCluster() []m.Cluster {
	clusters := []m.Cluster{}
	GCoreDB.Find(&clusters)

	return clusters
}

// GetCluster query for the cluster by given fuzzy clusterId
func GetClusterByFuzzyClusterId(clusterId string) *m.Cluster {
	query := "%" + strings.ToLower(clusterId) + "%"
	cluster := m.Cluster{}
	GCoreDB.Where("id LIKE ?", query).First(&cluster)
	if cluster.ID != "" {
		return &cluster
	}
	return nil
}

// GetClusterByIdentifier query for the cluster by given identifier, which is a random string generated when the
// cluster was created.
func GetClusterByIdentifier(clusterIdentifier string) *m.Cluster {
	cluster := m.Cluster{}
	GCoreDB.Where(&m.Cluster{Identifier: clusterIdentifier}).First(&cluster)
	if cluster.ID != "" {
		return &cluster
	}
	return nil
}

func CreateCluster(cluster *m.Cluster) error {
	// Generate a random identifier by default, prepend the clusterID to avoid name conflict
	if cluster.Identifier == "" {
		cluster.Identifier = strings.ToLower(cluster.ID) + "-" + uniuri.NewLen(16)
	}
	err := GCoreDB.Create(cluster).Error
	return err
}
