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

package app

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/modules"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	jsoniter "github.com/json-iterator/go"
	"github.com/parnurzeal/gorequest"
	"github.com/spf13/viper"
)

const (
	defaultNamespace     = "default"
	clusterServiceName   = "kubernetes"
	directConnectionMode = "direct"
)

const (
	// masterRole label
	masterRole = "node-role.kubernetes.io/master"
	// controlPlanRole label
	controlPlanRole = "node-role.kubernetes.io/control-plane"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// ClusterInfoParams parameters of cluster info
type ClusterInfoParams struct {
	ServerKey     string `json:"serverKey"`
	ClusterID     string `json:"clusterID"`
	ClientModule  string `json:"clientModule"`
	ServerAddress string `json:"serverAddress"`
	CaCertData    string `json:"caCertData"`
	UserToken     string `json:"userToken"`
	ClusterDomain string `json:"clusterDomain"`
}

func reportToBke(kubeCtx *KubeClientContext) { // nolint
	periodSync := viper.GetInt("agent.periodSync")
	clusterID := viper.GetString("cluster.id")
	monitorTicker := time.NewTicker(time.Duration(periodSync) * time.Second)
	defer monitorTicker.Stop()

	for {
		serverAddresses, err := kubeCtx.GetApiserverAddresses()
		if err != nil {
			blog.Errorf("Error getting apiserver addresses of cluster: %s", err.Error())
			// sleep a while to try again, avoid trying in loop
			time.Sleep(30 * time.Second)
			reportBcsKubeAgentReadiness(directConnectionMode, BCSKubeAgentStatesNotReady)
			continue
		}
		blog.Infof("apiserver addresses: %s", serverAddresses)

		bkeURL := getBkeAgentInfo()
		blog.Infof("bke-server url：%s", bkeURL)

		clusterInfoParams := ClusterInfoParams{
			ServerKey:     clusterID,
			ClusterID:     clusterID,
			ClientModule:  modules.BCSModuleKubeagent,
			ServerAddress: serverAddresses,
			CaCertData:    string(kubeCtx.GetRestConfig().CAData),
			UserToken:     kubeCtx.GetRestConfig().BearerToken,
		}

		var (
			handler = "clustermanagerReportCredentials"
			method  = "POST"
			start   = time.Now()
		)
		var request *gorequest.SuperAgent
		insecureSkipVerify := viper.GetBool("agent.insecureSkipVerify")
		if insecureSkipVerify {
			// NOCC:gas/tls(设计如此:此处需要跳过验证)
			request = gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: insecureSkipVerify}) // nolint
		} else {
			var tlsConfig *tls.Config
			caCrtFile := os.Getenv("CLIENT_CA")
			clientCrtFile := os.Getenv("CLIENT_CERT")
			clientKeyFile := os.Getenv("CLIENT_KEY")
			if len(clientCrtFile) == 0 && len(clientKeyFile) == 0 {
				tlsConfig, err = ssl.ClientTslConfVerityServer(caCrtFile)
			} else {
				tlsConfig, err = ssl.ClientTslConfVerity(caCrtFile, clientCrtFile, clientKeyFile, static.ClientCertPwd)
			}
			if err != nil {
				blog.Errorf("get client tls config failed, err %s", err.Error())
				reportBcsKubeAgentReadiness(directConnectionMode, BCSKubeAgentStatesNotReady)
				break
			}
			request = gorequest.New().TLSClientConfig(tlsConfig)
		}
		userToken := os.Getenv("USER_TOKEN")
		if len(userToken) == 0 {
			blog.Errorf("lost USER_TOKEN env parameter")
			panic("lost USER_TOKEN env parameter")
		}
		resp, respBody, errs := request.Put(bkeURL).
			Set("Authorization", "Bearer "+userToken).
			Send(clusterInfoParams).End()
		if len(errs) > 0 {
			blog.Errorf("unable to connect to the bke server: %s", errs[0].Error())
			reportBcsKubeAgentAPIMetrics(handler, method, FailConnect, start)
			// sleep a while to try again, avoid trying in loop
			time.Sleep(30 * time.Second)
			reportBcsKubeAgentReadiness(directConnectionMode, BCSKubeAgentStatesNotReady)
			continue
		}
		if resp.StatusCode >= 400 {
			reportBcsKubeAgentAPIMetrics(handler, method, fmt.Sprintf("%d", resp.StatusCode), start)
			reportBcsKubeAgentReadiness(directConnectionMode, BCSKubeAgentStatesNotReady)
			blog.Errorf("resp code %d, respBody %s", resp.StatusCode, respBody)
		} else {
			codeName := json.Get([]byte(respBody), "code").ToInt()
			message := json.Get([]byte(respBody), "message").ToString()
			if codeName != 0 {
				reportBcsKubeAgentReadiness(directConnectionMode, BCSKubeAgentStatesNotReady)
				blog.Errorf(
					"Error updating cluster credential to bke, response code: %s, response message: %s",
					codeName,
					message,
				)
			}
			reportBcsKubeAgentReadiness(directConnectionMode, BCSKubeAgentStatesReady)
			reportBcsKubeAgentAPIMetrics(handler, method, fmt.Sprintf("%d", codeName), start)
		}

		select { // nolint
		case <-monitorTicker.C:
		}
	}
}

func getBkeAgentInfo() string {
	bkeServerAddress := viper.GetString("bke.serverAddress")
	bkeReportPath := viper.GetString("bke.report-path")
	bkeURL := bkeServerAddress + bkeReportPath

	return bkeURL
}
