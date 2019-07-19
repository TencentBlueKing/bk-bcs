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
	m "bk-bcs/bcs-services/bcs-api/pkg/models"
)

// GetCredentials query for clusterCredentials by clusterId
func GetCredentials(clusterId string) *m.ClusterCredentials {
	credentials := m.ClusterCredentials{}
	GCoreDB.Where(&m.ClusterCredentials{ClusterId: clusterId}).First(&credentials)
	if credentials.ID != 0 {
		return &credentials
	}
	return nil
}

// SaveCredentials saves the current cluster credentials
func SaveCredentials(clusterId, serverAddresses, caCertData, userToken, clusterDomain string) error {
	var credentials m.ClusterCredentials
	// Create or update, source: https://github.com/jinzhu/gorm/issues/1307
	dbScoped := GCoreDB.Where(m.ClusterCredentials{ClusterId: clusterId}).Assign(
		m.ClusterCredentials{
			ServerAddresses: serverAddresses,
			CaCertData:      caCertData,
			UserToken:       userToken,
			ClusterDomain:   clusterDomain,
		},
	).FirstOrCreate(&credentials)
	return dbScoped.Error
}
