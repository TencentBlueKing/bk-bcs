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

import (
	"strings"
	"time"
)

const (
	ClusterProviderPlain = iota + 1
	ClusterProviderBCS
	ClusterProviderFixture

	ClusterIdPrefixPlain = "plain-"
	ClusterIdPrefixBCS   = "bcs-"
)

type Cluster struct {
	ID string `json:"id" gorm:"primary_key"`
	// Provider field shows that which provider did this cluster belongs to, provider will determine the authorization
	// procedure when someone tries to interact with this cluster, possible values:
	//
	//   - plain: cluster with this provider can only be viewed or updated by its creator
	//   - bcs: user cann't interact with this cluster unless the "blueking PERM" service said "yes"
	Provider uint `json:"provider"`
	// CreatorId is the user_id of creator
	CreatorId uint `json:"creator_id"`
	// Identifier is a random string, it can be used for apiserver proxy tunnel URL address concatenation
	Identifier string    `json:"identifier" gorm:"size:128"`
	CreatedAt  time.Time `json:"created_at"`
}

type ClusterCredentials struct {
	ID        uint   `gorm:"primary_key"`
	ClusterId string `gorm:"unique;not null"`
	// ServerAddresses is all available apiserver addresses, separated by ";", for example: "https//x.com;http://y.com"
	ServerAddresses string `gorm:"size:2048"`
	CaCertData      string `gorm:"size:4096"`
	UserToken       string `gorm:"size:2048"`
	ClusterDomain   string `gorm:"size:2048"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// GetServerAddressesList returns the apisrever list which was separated by ";"
func (c *ClusterCredentials) GetServerAddressesList() []string {
	if strings.Trim(c.ServerAddresses, " ") == "" {
		return []string{}
	}
	return strings.Split(c.ServerAddresses, ";")
}

// RegisterToken was issued when one cluster agent want to register it's credential informations to bke-server
type RegisterToken struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	ClusterId string    `json:"cluster_id" gorm:"unique;not null"`
	Token     string    `json:"token" gorm:"size:256"`
	CreatedAt time.Time `json:"created_at"`
}

type WsClusterCredentials struct {
	ID            uint   `gorm:"primary_key"`
	ServerKey     string `gorm:"unique;not null"`
	ClientModule  string `gorm:"not null"`
	ServerAddress string `gorm:"size:2048"`
	CaCertData    string `gorm:"size:4096"`
	UserToken     string `gorm:"size:2048"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
