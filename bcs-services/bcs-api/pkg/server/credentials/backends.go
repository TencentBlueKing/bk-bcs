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

package credentials

import (
	"encoding/base64"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/config"
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
)

type CredentialBackend interface {
	GetClusterByIdentifier(clusterIdentifier string) (*m.Cluster, error)
	GetCredentials(clusterId string) (*m.ClusterCredentials, error)
}

// A backend which loads cluster credentials from config file
type FixtureCredentialBackend struct {
	credentials map[string]*m.ClusterCredentials
}

var GFixtureCredentialBackend = &FixtureCredentialBackend{
	credentials: make(map[string]*m.ClusterCredentials),
}

func (l *FixtureCredentialBackend) ExtractCredentialsFixtures() error {
	isEnabled := config.ClusterCredentialsFixtures.Enabled
	if !isEnabled {
		return nil
	}
	for _, clusterCred := range config.ClusterCredentialsFixtures.Credentials {
		// Only "service_account" type is supported at the moment
		if clusterCred.Type != "service_account" {
			continue
		}
		if clusterCred.ClusterID == "" {
			return fmt.Errorf("cluster_id must not be empty")
		}

		// Decode original cert and token
		caCert, err := base64.StdEncoding.DecodeString(clusterCred.CaCert)
		if err != nil {
			return fmt.Errorf("ca_cert must be base64 encoded string")
		}
		token, err := base64.StdEncoding.DecodeString(clusterCred.Token)
		if err != nil {
			return fmt.Errorf("token must be base64 encoded string")
		}

		clusterCredentials := &m.ClusterCredentials{
			ClusterId:       clusterCred.ClusterID,
			ServerAddresses: clusterCred.Server,
			CaCertData:      string(caCert),
			UserToken:       string(token),
		}
		l.credentials[clusterCred.ClusterID] = clusterCredentials
	}
	return nil
}

// GetClusterByIdentifier, for fixture backend, it will only try to use the identifier as clusterId to find the
// Cluster.
func (l *FixtureCredentialBackend) GetClusterByIdentifier(clusterIdentifier string) (*m.Cluster, error) {
	// Use Identifier as ID
	result := l.credentials[clusterIdentifier]
	if result == nil {
		return nil, fmt.Errorf("cluster not found with identifier=%s", clusterIdentifier)
	}
	return &m.Cluster{
		ID:       clusterIdentifier,
		Provider: m.ClusterProviderFixture,
	}, nil
}

// GetCredentials get a credential by clusterId
func (l *FixtureCredentialBackend) GetCredentials(clusterId string) (*m.ClusterCredentials, error) {
	result := l.credentials[clusterId]
	if result == nil {
		return nil, fmt.Errorf("credentials not found for %s", clusterId)
	}
	return result, nil
}

// A backend which loads credentials from database
type DatabaseCrendentialBackend struct{}

var GDatabaseCrendentialBackend = &DatabaseCrendentialBackend{}

// GetClusterByIdentifier, for database backend, it will query the database using given clusterIdenfier.
func (l *DatabaseCrendentialBackend) GetClusterByIdentifier(clusterIdentifier string) (*m.Cluster, error) {
	cluster := sqlstore.GetClusterByIdentifier(clusterIdentifier)
	return cluster, nil
}

// GetCredentials get a credential from clusterId
func (l *DatabaseCrendentialBackend) GetCredentials(clusterId string) (*m.ClusterCredentials, error) {
	credentials := sqlstore.GetCredentials(clusterId)
	return credentials, nil
}
