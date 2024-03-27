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

// Package api xxx
package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
	iamModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// GetInternalClusterKubeConfig get cce cluster kebeconfig
func GetInternalClusterKubeConfig(client *CceClient, clusterId string) (string, error) {
	req := model.CreateKubernetesClusterCertRequest{
		ClusterId: clusterId, // 集群ID，可在CCE管理控制台中查看
		Body: &model.CertDuration{
			Duration: int32(-1), // 集群证书有效时间，单位为天，最小值为1，最大值为10950(30*365，1年固定计365天，忽略闰年影响)；若填-1则为最大值30年。
		},
	}

	rsp, err := client.CreateKubernetesClusterCert(&req)
	if err != nil {
		return "", err
	}

	currentContext := "internal"
	clusters := make([]model.Clusters, 0)
	contexts := make([]model.Contexts, 0)

	for _, v := range *rsp.Clusters {
		if *v.Name == "internalCluster" {
			clusters = append(clusters, v)
		}
	}

	if len(clusters) == 0 {
		return "", fmt.Errorf("internal cluster not found")
	}

	for _, v := range *rsp.Contexts {
		if *v.Name == "internal" {
			contexts = append(contexts, v)
		}
	}

	if len(contexts) == 0 {
		return "", fmt.Errorf("internal context not found")
	}

	kubeCfg := &model.CreateKubernetesClusterCertResponse{
		Kind:           rsp.Kind,
		ApiVersion:     rsp.ApiVersion,
		Preferences:    rsp.Preferences,
		Clusters:       &clusters,
		Users:          rsp.Users,
		Contexts:       &contexts,
		CurrentContext: &currentContext,
		PortID:         rsp.PortID,
	}

	bt, err := json.Marshal(kubeCfg)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bt), nil
}

// GetClusterKubeConfig get cce cluster kebeconfig
func GetClusterKubeConfig(client *CceClient, clusterId string) (string, error) {
	req := model.CreateKubernetesClusterCertRequest{
		ClusterId: clusterId, // 集群ID，可在CCE管理控制台中查看
		Body: &model.CertDuration{
			Duration: int32(-1), // 集群证书有效时间，单位为天，最小值为1，最大值为10950(30*365，1年固定计365天，忽略闰年影响)；若填-1则为最大值30年。
		},
	}

	rsp, err := client.CreateKubernetesClusterCert(&req)
	if err != nil {
		return "", err
	}

	kubeCfg := &model.CreateKubernetesClusterCertResponse{}
	if len(*rsp.Clusters) == 1 {
		kubeCfg = rsp
	} else if len(*rsp.Clusters) > 1 && *rsp.CurrentContext == "external" {
		curContext := "externalTLSVerify"
		clusters := make([]model.Clusters, 0)
		contexts := make([]model.Contexts, 0)

		for _, v := range *rsp.Clusters {
			if *v.Name == "externalClusterTLSVerify" {
				clusters = append(clusters, v)
			}
		}

		for _, v := range *rsp.Contexts {
			if *v.Name == "externalTLSVerify" {
				contexts = append(contexts, v)
			}
		}
		kubeCfg = &model.CreateKubernetesClusterCertResponse{
			Kind:           rsp.Kind,
			ApiVersion:     rsp.ApiVersion,
			Preferences:    rsp.Preferences,
			Clusters:       &clusters,
			Users:          rsp.Users,
			Contexts:       &contexts,
			CurrentContext: &curContext,
			PortID:         rsp.PortID,
		}
	}

	bt, err := json.Marshal(kubeCfg)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bt), nil
}

// GetProjectIDByRegion get project ID by region
func GetProjectIDByRegion(opt *cloudprovider.CommonOption) (string, error) {
	client, err := GetIamClient(opt)
	if err != nil {
		return "", err
	}

	req := iamModel.KeystoneListProjectsRequest{Name: &opt.Region}
	rsp, err := client.KeystoneListProjects(&req)
	if err != nil {
		return "", err
	}

	if len(*rsp.Projects) == 0 {
		return "", fmt.Errorf("project not found")
	} else if len(*rsp.Projects) > 1 {
		return "", fmt.Errorf("the number of project is greater than one")
	}

	return (*rsp.Projects)[0].Id, nil
}
