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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
)

// GetCredentials query for clusterCredentials by clusterId
func GetCredentials(clusterID string) *models.BcsClusterCredential {
	credential := models.BcsClusterCredential{}
	GCoreDB.Where(&models.BcsClusterCredential{ClusterId: clusterID}).First(&credential)
	if credential.ID != 0 {
		return &credential
	}
	return nil
}

// SaveCredentials saves the current cluster credentials
func SaveCredentials(clusterID, serverAddresses, caCertData, userToken, clusterDomain string) error {
	var credentials models.BcsClusterCredential
	// Create or update, source: https://github.com/jinzhu/gorm/issues/1307
	dbScoped := GCoreDB.Where(models.BcsClusterCredential{ClusterId: clusterID}).Assign(
		models.BcsClusterCredential{
			ServerAddresses: serverAddresses,
			CaCertData:      caCertData,
			UserToken:       userToken,
			ClusterDomain:   clusterDomain,
		},
	).FirstOrCreate(&credentials)
	return dbScoped.Error
}

// ListCredentials list cluster credentials
func ListCredentials() []models.BcsClusterCredential {
	var credentials []models.BcsClusterCredential
	GCoreDB.Find(&credentials)

	return credentials
}

// SaveWsCredentials saves the credentials of cluster registered by websocket
func SaveWsCredentials(serverKey, clientModule, serverAddress, caCertData, userToken string) error {
	var credentials models.BcsWsClusterCredentials
	// Create or update, source: https://github.com/jinzhu/gorm/issues/1307
	dbScoped := GCoreDB.Where(models.BcsWsClusterCredentials{ServerKey: serverKey}).Assign(
		models.BcsWsClusterCredentials{
			ClientModule:  clientModule,
			ServerAddress: serverAddress,
			CaCertData:    caCertData,
			UserToken:     userToken,
		},
	).FirstOrCreate(&credentials)
	return dbScoped.Error
}

// GetWsCredentials query for clusterCredentials of cluster registered by websocket
func GetWsCredentials(serverKey string) *models.BcsWsClusterCredentials {
	credentials := models.BcsWsClusterCredentials{}
	GCoreDB.Where(&models.BcsWsClusterCredentials{ServerKey: serverKey}).First(&credentials)
	if credentials.ID != 0 {
		return &credentials
	}
	return nil
}

// DelWsCredentials delete ws credentials
func DelWsCredentials(serverKey string) {
	credentials := models.BcsWsClusterCredentials{}
	GCoreDB.Where(&models.BcsWsClusterCredentials{ServerKey: serverKey}).Delete(&credentials)
}

// GetWsCredentialsByClusterId get ws credential by clusterID
func GetWsCredentialsByClusterId(clusterID string) []*models.BcsWsClusterCredentials {
	var credentials []*models.BcsWsClusterCredentials
	query := clusterID + "-%"
	GCoreDB.Where("server_key LIKE ?", query).Find(&credentials)
	return credentials
}
