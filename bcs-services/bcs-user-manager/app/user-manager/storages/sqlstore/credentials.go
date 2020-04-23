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
	"bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
)

// GetCredentials query for clusterCredentials by clusterId
func GetCredentials(clusterId string) *models.BcsClusterCredential {
	credential := models.BcsClusterCredential{}
	GCoreDB.Where(&models.BcsClusterCredential{ClusterId: clusterId}).First(&credential)
	if credential.ID != 0 {
		return &credential
	}
	return nil
}

// SaveCredentials saves the current cluster credentials
func SaveCredentials(clusterId, serverAddresses, caCertData, userToken, clusterDomain string) error {
	var credentials models.BcsClusterCredential
	// Create or update, source: https://github.com/jinzhu/gorm/issues/1307
	dbScoped := GCoreDB.Where(models.BcsClusterCredential{ClusterId: clusterId}).Assign(
		models.BcsClusterCredential{
			ServerAddresses: serverAddresses,
			CaCertData:      caCertData,
			UserToken:       userToken,
			ClusterDomain:   clusterDomain,
		},
	).FirstOrCreate(&credentials)
	return dbScoped.Error
}

func ListCredentials() []models.BcsClusterCredential {
	var credentials []models.BcsClusterCredential
	GCoreDB.Find(&credentials)

	return credentials
}
