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

package tke

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/external-cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/external-cluster/tke/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
)

const (
	// TkeSdkToGetCredentials handler for get credential
	TkeSdkToGetCredentials = "DescribeClusterSecurityInfo"
	// HTTPScheme https
	HTTPScheme = "https://"
	// TkeClusterPort cluster port
	TkeClusterPort = ":443"
)

type tkeCluster struct {
	ClusterID        string
	TkeClusterID     string
	TkeClusterRegion string
}

// Client cluster
type Client struct {
	*common.Client
}

// Response for resp
type Response struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	CodeDesc string `json:"codeDesc"`
}

// DescribeClusterSecurityInfoArgs clusterID args
type DescribeClusterSecurityInfoArgs struct {
	ClusterID string `qcloud_arg:"clusterId"`
}

// BindMasterVipLoadBalancerArgs vip args
type BindMasterVipLoadBalancerArgs struct {
	ClusterID string `qcloud_arg:"clusterId"`
	SubnetID  string `qcloud_arg:"subnetId"`
}

// BindMasterVipLoadBalanceResponse xxx
type BindMasterVipLoadBalanceResponse struct {
	Response
	Data interface{}
}

// GetMasterVipArgs args
type GetMasterVipArgs struct {
	ClusterID string `qcloud_arg:"clusterId"`
}

// GetMasterVipResponse xxx
type GetMasterVipResponse struct {
	Response
	Data GetMasterVipRespData
}

// GetMasterVipRespData xxx
type GetMasterVipRespData struct {
	Status string `json:"status"`
}

// DescribeClusterSecurityInfoRespData cluster security info resp data
type DescribeClusterSecurityInfoRespData struct {
	UserName                string `json:"userName"`
	Domain                  string `json:"domain"`
	CertificationAuthority  string `json:"certificationAuthority"`
	PgwEndpoint             string `json:"pgwEndpoint"`
	ClusterExternalEndpoint string `json:"clusterExternalEndpoint"`
	Password                string `json:"password"`
}

// DescribeClusterSecurityInfoResponse response
type DescribeClusterSecurityInfoResponse struct {
	Response
	Data DescribeClusterSecurityInfoRespData `json:"data"`
}

// NewTkeCluster init tkeCluster client
func NewTkeCluster(clusterId, tkeClusterId, tkeClusterRegion string) external_cluster.ExternalCluster {
	return &tkeCluster{
		ClusterID:        clusterId,
		TkeClusterID:     tkeClusterId,
		TkeClusterRegion: tkeClusterRegion,
	}
}

// SyncClusterCredentials sync cluster credentials
func (t *tkeCluster) SyncClusterCredentials() error {
	tkeClient, err := NewClient(t.TkeClusterRegion, "GET")
	if err != nil {
		return fmt.Errorf("error when creating tke client: %s", err.Error())
	}

	args := DescribeClusterSecurityInfoArgs{
		ClusterID: t.TkeClusterID,
	}
	response := &DescribeClusterSecurityInfoResponse{}
	err = tkeClient.Invoke(TkeSdkToGetCredentials, args, response)
	if err != nil {
		return fmt.Errorf("error when invoking tke api %s: %s", TkeSdkToGetCredentials, err.Error())
	}
	if response.Code != 0 {
		return fmt.Errorf("%s cluster %s failed, codeDesc: %s, message: %s", TkeSdkToGetCredentials, t.TkeClusterID, response.CodeDesc, response.Message)
	}

	if response.Data.PgwEndpoint == "" || response.Data.Domain == "" {
		return fmt.Errorf("BindMasterVipLoadBalancer failed, pgwEndpoint or domain nil")
	}

	serverAddress := HTTPScheme + response.Data.PgwEndpoint + TkeClusterPort
	clusterDomainURL := HTTPScheme + response.Data.Domain + "/"
	err = sqlstore.SaveCredentials(t.ClusterID, serverAddress, response.Data.CertificationAuthority, response.Data.Password, clusterDomainURL)
	if err != nil {
		return fmt.Errorf("error when updating external cluster credentials to db: %s", err.Error())
	}
	return nil
}

// NewClient init tkeCluster client
func NewClient(tkeClusterRegion, method string) (*Client, error) {
	tkeSecretID := config.Tke.SecretId
	tkeSecretKey := config.Tke.SecretKey
	tkeCcsHost := config.Tke.CcsHost
	tkeCcsPath := config.Tke.CcsPath
	if tkeSecretID == "" || tkeSecretKey == "" || tkeCcsHost == "" || tkeCcsPath == "" {
		return nil, fmt.Errorf("tke conf invalid")
	}
	credential := common.Credential{
		SecretID:  tkeSecretID,
		SecretKey: tkeSecretKey,
	}

	opts := common.Opts{
		Region: tkeClusterRegion,
		Host:   tkeCcsHost,
		Path:   tkeCcsPath,
		Method: method,
	}

	client, err := common.NewClient(credential, opts)
	if err != nil {
		return &Client{}, err
	}
	return &Client{client}, nil
}
