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

package models

import "time"

// BcsCluster table
type BcsCluster struct {
	ID               string    `json:"id" gorm:"primary_key"`
	ClusterType      uint      `json:"cluster_type"`
	TkeClusterId     string    `json:"tke_cluster_id"`
	TkeClusterRegion string    `json:"tke_cluster_region"`
	CreatorId        uint      `json:"creator_id"`
	CreatedAt        time.Time `json:"created_at"`
}

// BcsRegisterToken was issued when one cluster agent want to register it's credential informations to bke-server
type BcsRegisterToken struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	ClusterId string    `json:"cluster_id" gorm:"unique;not null"`
	Token     string    `json:"token" gorm:"size:256"`
	CreatedAt time.Time `json:"created_at"`
}

// BcsClusterCredential table
type BcsClusterCredential struct {
	ID              uint      `json:"id" gorm:"primary_key"`
	ClusterId       string    `json:"cluster_id" gorm:"unique;not null"`
	ServerAddresses string    `json:"server_addresses" gorm:"size:2048"`
	CaCertData      string    `json:"ca_cert_data" gorm:"size:4096"`
	UserToken       string    `json:"user_token" gorm:"size:2048"`
	ClusterDomain   string    `json:"cluster_domain" gorm:"size:2048"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// BcsWsClusterCredentials table
type BcsWsClusterCredentials struct {
	ID            uint   `gorm:"primary_key"`
	ServerKey     string `gorm:"unique;not null"`
	ClientModule  string `gorm:"not null"`
	ServerAddress string `gorm:"size:2048"`
	CaCertData    string `gorm:"size:4096"`
	UserToken     string `gorm:"size:2048"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
