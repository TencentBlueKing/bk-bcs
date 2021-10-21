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
	"time"
)

func GetUserClusterPerm(user *m.User, cluster *m.Cluster, name string) *m.UserClusterPermission {
	var result m.UserClusterPermission
	GCoreDB.Where(m.UserClusterPermission{UserID: user.ID, ClusterID: cluster.ID, Name: name}).First(&result)
	if result.ID == 0 {
		return nil
	}
	return &result
}

// SaveUserClusterPerm saves a user cluster permission to database, it will update the `UpdatedAt` filed to
// current timestamp.
func SaveUserClusterPerm(backend int, user *m.User, cluster *m.Cluster, name string, isActive bool) error {
	var result m.UserClusterPermission
	// Create or update, source: https://github.com/jinzhu/gorm/issues/1307
	dbScoped := GCoreDB.Where(m.UserClusterPermission{
		Backend:   backend,
		UserID:    user.ID,
		ClusterID: cluster.ID,
		Name:      name,
	}).Assign(
		m.UserClusterPermission{
			IsActive:  isActive,
			UpdatedAt: time.Now(),
		},
	).FirstOrCreate(&result)
	return dbScoped.Error
}
