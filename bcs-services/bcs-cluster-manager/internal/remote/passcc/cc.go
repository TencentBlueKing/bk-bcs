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

// Package passcc xxx
package passcc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/utils"
	iutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var (
	defaultTimeOut   = time.Second * 60
	errServerNotInit = errors.New("server not inited")
)

// CClient global client
var CClient *ClientConfig

// SetCCClient set pass-cc client
func SetCCClient(options Options) error {
	if !options.Enable {
		CClient = nil
		return nil
	}
	cli := NewCCClient(options)

	CClient = cli
	return nil
}

// GetCCClient get pass-cc client
func GetCCClient() *ClientConfig {
	return CClient
}

// NewCCClient for init cc client
func NewCCClient(opt Options) *ClientConfig {
	cli := &ClientConfig{
		server:    opt.Server,
		appCode:   opt.AppCode,
		appSecret: opt.AppSecret,
		debug:     opt.Debug,
	}
	return cli
}

// Options opts parameter
type Options struct {
	// Server auth address
	Server string
	// AppCode app code
	AppCode string
	// AppSecret app secret
	AppSecret string
	// Enable enable feature
	Enable bool
	// Debug http debug
	Debug bool
}

// ClientConfig pass-cc client
type ClientConfig struct {
	server string

	appCode   string
	appSecret string
	debug     bool
}

// CreatePassCCClusterSnapshoot register cluster scapshoot to pass-cc
func (cc *ClientConfig) CreatePassCCClusterSnapshoot(cluster *proto.Cluster) error {
	if cc == nil {
		return errServerNotInit
	}
	var (
		_    = "CreatePassCCClusterSnapshoot"
		path = fmt.Sprintf("/v1/clusters/%s/cluster_config/", cluster.ClusterID)
	)

	// get access_token
	token, err := cc.getAccessToken(nil)
	if err != nil {
		blog.Errorf("CreatePassCCClusterSnapshoot call getAccessToken failed: %v", err)
		return err
	}

	// default field
	clusterReq := cc.transClusterToClusterSnap(cluster)
	var (
		url  = cc.server + path
		resp = &CommonResp{}
	)

	result, body, errs := gorequest.New().Timeout(defaultTimeOut).Post(url).
		Set("X-Bkapi-Authorization", token).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		SetDebug(true).
		Send(clusterReq).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call api CreatePassCCClusterSnapshoot failed: %v", errs[0])
		return errs[0]
	}

	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("call CreatePassCCClusterSnapshoot API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return errMsg
	}

	blog.Infof("CreatePassCCClusterSnapshoot[%s] successful", cluster.ClusterID)
	return nil
}

// DeletePassCCCluster delete cluster in pass-cc
func (cc *ClientConfig) DeletePassCCCluster(projectID, clusterID string) error {
	if cc == nil {
		return errServerNotInit
	}
	var (
		_    = "DeletePassCCCluster"
		path = fmt.Sprintf("/projects/%s/clusters/%s/", projectID, clusterID)
	)

	// get access_token
	token, err := cc.getAccessToken(nil)
	if err != nil {
		blog.Errorf("DeletePassCCCluster call getAccessToken failed: %v", err)
		return err
	}

	var (
		url  = cc.server + path
		resp = &CommonResp{}
	)

	result, body, errs := gorequest.New().Timeout(defaultTimeOut).Delete(url).
		Set("X-Bkapi-Authorization", token).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		SetDebug(true).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call api DeletePassCCCluster failed: %v", errs[0])
		return errs[0]
	}

	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("call DeletePassCCCluster API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return errMsg
	}

	blog.Infof("DeletePassCCCluster[%s] successful", clusterID)
	return nil
}

// GetProjectSharedNamespaces get namespaces in pass-cc
func (cc *ClientConfig) GetProjectSharedNamespaces(
	projectID, clusterID string, auth *auth.ClientAuth) ([]ProjectNamespace, error) {
	if cc == nil {
		return nil, errServerNotInit
	}
	var (
		_    = "GetProjectSharedNamespaces"
		path = fmt.Sprintf("/projects/%s/clusters/%s/namespaces", projectID, clusterID)
	)

	// get access_token
	token, err := cc.getAccessToken(auth)
	if err != nil {
		blog.Errorf("GetProjectSharedNamespaces call getAccessToken failed: %v", err)
		return nil, err
	}

	var (
		url  = cc.server + path
		req  = &GetProjectsNamespaces{DesireAllData: 1}
		resp = &GetProjectsNamespacesResp{}
	)

	result, body, errs := gorequest.New().Timeout(defaultTimeOut).Get(url).
		Set("X-Bkapi-Authorization", token).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		SetDebug(true).
		Send(req).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call api GetProjectSharedNamespaces failed: %v", errs[0])
		return nil, errs[0]
	}

	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("call GetProjectSharedNamespaces API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return nil, errMsg
	}

	blog.Infof("GetProjectSharedNamespaces[%s:%s] successful", projectID, clusterID)
	return resp.Data.Results, nil
}

// CreatePassCCCluster register cluster to pass-cc
func (cc *ClientConfig) CreatePassCCCluster(cluster *proto.Cluster) error {
	if cc == nil {
		return errServerNotInit
	}
	var (
		_    = "CreatePassCCCluster"
		path = fmt.Sprintf("/projects/%s/clusters/", cluster.ProjectID)
	)

	// get access_token
	token, err := cc.getAccessToken(nil)
	if err != nil {
		blog.Errorf("CreatePassCCCluster call getAccessToken failed: %v", err)
		return err
	}

	// default field
	clusterReq := cc.transCMClusterToCC(cluster)
	var (
		url  = cc.server + path
		resp = &CommonResp{}
	)

	result, body, errs := gorequest.New().Timeout(defaultTimeOut).Post(url).
		Set("X-Bkapi-Authorization", token).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		SetDebug(true).
		Send(clusterReq).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call api CreatePassCCCluster failed: %v", errs[0])
		return errs[0]
	}

	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("call CreatePassCCCluster API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return errMsg
	}

	blog.Infof("CreatePassCCCluster[%s] successful", cluster.ClusterID)
	return nil
}

// UpdatePassCCCluster update cluster info to pass-cc
func (cc *ClientConfig) UpdatePassCCCluster(cluster *proto.Cluster) error {
	if cc == nil {
		return errServerNotInit
	}
	var (
		_    = "UpdatePassCCCluster"
		path = fmt.Sprintf("/projects/%s/clusters/%s/", cluster.ProjectID, cluster.ClusterID)
	)

	// get access_token
	token, err := cc.getAccessToken(nil)
	if err != nil {
		blog.Errorf("UpdatePassCCCluster call getAccessToken failed: %v", err)
		return err
	}

	// default field
	clusterReq := cc.transCMClusterToCC(cluster)
	var (
		url  = cc.server + path
		resp = &CommonResp{}
	)

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Put(url).
		Set("X-Bkapi-Authorization", token).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		SetDebug(true).
		Send(clusterReq).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call api UpdatePassCCCluster failed: %v", errs[0])
		return errs[0]
	}

	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("call UpdatePassCCCluster API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return errMsg
	}

	blog.Infof("UpdatePassCCCluster[%s] successful", cluster.ClusterID)
	return nil
}

func (cc *ClientConfig) getAccessToken(clientAuth *auth.ClientAuth) (string, error) {
	if cc == nil {
		return "", errServerNotInit
	}

	var (
		appToken string
		err      error
	)

	if clientAuth != nil {
		appToken, err = clientAuth.GetAccessToken(utils.BkAppUser{
			BkAppCode:   cc.appCode,
			BkAppSecret: cc.appSecret,
		})
	} else {
		appToken, err = auth.GetAccessClient().GetAccessToken(utils.BkAppUser{
			BkAppCode:   cc.appCode,
			BkAppSecret: cc.appSecret,
		})
	}
	if err != nil {
		return "", err
	}
	token := utils.BkAccessToken{AccessToken: appToken}

	tokenStr, _ := json.Marshal(token)
	return string(tokenStr), nil
}

func (cc *ClientConfig) transClusterToClusterSnap(cls *proto.Cluster) *CreateClusterConfParams {
	masterIPs := make([]string, 0)
	for ip := range cls.Master {
		masterIPs = append(masterIPs, ip)
	}

	clusterSnapInfo := &ClusterSnapShootInfo{
		Regions:      cls.Region,
		ClusterID:    cls.ClusterID,
		MasterIPList: masterIPs,
		VpcID:        cls.VpcID,
		SystemDataID: 21449,
		ClusterCIDRSettings: ClusterCIDRInfo{
			ClusterCIDR:          cls.GetNetworkSettings().GetClusterIPv4CIDR(),
			MaxNodePodNum:        cls.GetNetworkSettings().GetMaxNodePodNum(),
			MaxClusterServiceNum: cls.GetNetworkSettings().GetMaxServiceNum(),
		},
		ClusterType: cls.ClusterType,
		ClusterBasicSettings: ClusterBasicInfo{
			ClusterOS:      cls.GetClusterBasicSettings().GetOS(),
			ClusterVersion: cls.GetClusterBasicSettings().GetVersion(),
			ClusterName:    cls.ClusterName,
		},
		ClusterAdvancedSettings: ClusterAdvancedInfo{
			IPVS: cls.GetClusterAdvanceSettings().GetIPVS(),
		},
		NetWorkType:    cls.NetworkType,
		EsbURL:         defaultEsbURL,
		WebhookImage:   defaultWebhookImage,
		PrivilegeImage: defaultPrivilegeImage,
		VersionName:    cls.GetClusterBasicSettings().GetVersion(),
		Version:        cls.GetClusterBasicSettings().GetVersion(),
		ClusterVersion: cls.GetClusterBasicSettings().GetVersion(),
		ControlIP: func() string {
			if len(masterIPs) > 0 {
				return masterIPs[0]
			}
			return ""
		}(),
		MasterIPs:      masterIPs,
		Env:            cls.Environment,
		ProjectName:    cls.ProjectID,
		ProjectCode:    cls.ProjectID,
		AreaName:       cls.Region,
		ExtraClusterID: cls.SystemID,
	}

	conf, err := json.Marshal(clusterSnapInfo)
	if err != nil {
		blog.Errorf("transClusterToClusterSnap marshal clusterSnapInfo failed: %v", err)
		return &CreateClusterConfParams{
			Creator:   cls.Creator,
			ClusterID: cls.ClusterID,
			Configure: "",
		}
	}

	return &CreateClusterConfParams{
		Creator:   cls.Creator,
		ClusterID: cls.ClusterID,
		Configure: string(conf),
	}
}

func (cc *ClientConfig) transCMClusterToCC(cluster *proto.Cluster) *ClusterParamsRequest {
	var (
		areaID = 1
	)

	if v, ok := prodAreaCode[cluster.Region]; ok && strings.Contains(cc.server, "prod") {
		areaID = v
	} else if v, ok := testAreaCode[cluster.Region]; ok {
		areaID = v
	}

	masterIPs := make([]ManagerMasters, 0)
	if len(cluster.Master) == 0 {
		masterIPs = append(masterIPs, ManagerMasters{InnerIP: iutils.FakeIPV4Addr()})
	} else {
		for ip := range cluster.Master {
			masterIPs = append(masterIPs, ManagerMasters{InnerIP: ip})
		}
	}

	desc := cluster.Description
	if len(desc) == 0 {
		desc = cluster.ClusterID
	}

	return &ClusterParamsRequest{
		ClusterID:          cluster.ClusterID,
		ClusterName:        cluster.ClusterName,
		ClusterDescription: desc,
		AreaID:             areaID,
		VpcID:              cluster.VpcID,
		Env:                cluster.Environment,
		MasterIPs:          masterIPs,
		NeedNAT:            true,
		Version:            cluster.GetClusterBasicSettings().GetVersion(),
		NetworkType:        cluster.NetworkType,
		Coes:               "tke",
		KubeProxyMode:      "ipvs",
		Creator:            cluster.Creator,
		Type:               "tke",
		ExtraClusterID:     cluster.SystemID,
		State:              State,
		Status:             Status,
	}
}
