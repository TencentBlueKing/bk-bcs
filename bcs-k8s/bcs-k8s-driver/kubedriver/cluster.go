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

package kubedriver

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-driver/kubedriver/options"

	"github.com/parnurzeal/gorequest"
)

type AccessTokenRequestBody struct {
	EnvName    string `json:"env_name"`
	AppCode    string `json:"app_code"`
	AppSecret  string `json:"app_secret"`
	IDProvider string `json:"id_provider"`
	GrantType  string `json:"grant_type"`
}

type APIGWAuthHeaders struct {
	AccessToken string `json:"access_token"`
}

func GetClusterID(o *options.KubeDriverServerOptions) (clusterID string, err error) {
	if o.Environment == "develop" {
		return "driver-debug-clusterID", nil
	}
	if o.CustomClusterID != "" {
		return o.CustomClusterID, nil
	}
	err = o.GetClusterKeeperAddr()
	if nil != err {
		return "", fmt.Errorf("get cluster keeper  api server addr failed. reason: %v", err)
	}

	goReq := gorequest.New()
	if o.NeedClusterTLSConfig() {
		clusterTLSConfig, err := o.ClusterClientTLS.ToConfigObj()
		if err != nil {
			return "", fmt.Errorf("config cluster keeper tls failed, reason: %v", err)
		}
		goReq = goReq.TLSClientConfig(clusterTLSConfig)
	}

	generatedUrl := fmt.Sprintf("%s%s", o.ClusterKeeperUrl, "/bcsclusterkeeper/v1/cluster/id/byip")
	resp, respBody, errs := goReq.Get(generatedUrl).Query(map[string]string{"ip": o.HostIP}).End()
	if len(errs) != 0 {
		return "", errs[0]
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("cluster keeper request error, status code: %d, url: %s", resp.StatusCode, o.ClusterKeeperUrl)
	}

	clusterID = json.Get([]byte(respBody), "data").Get("clusterID").ToString()
	return clusterID, nil
}
