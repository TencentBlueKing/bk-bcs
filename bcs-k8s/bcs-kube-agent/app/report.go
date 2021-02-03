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

package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"

	"github.com/json-iterator/go"
	"github.com/parnurzeal/gorequest"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	defaultNamespace   = "default"
	clusterServiceName = "kubernetes"
	kubeAgentModule    = "kube-agent"
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

func reportToBke(kubeClient *kubernetes.Clientset, cfg *rest.Config) {
	periodSync := viper.GetInt("agent.periodSync")
	clusterID := viper.GetString("cluster.id")
	monitorTicker := time.NewTicker(time.Duration(periodSync) * time.Second)
	defer monitorTicker.Stop()
	for {
		serverAddresses, err := getApiserverAdresses(kubeClient)
		if err != nil {
			blog.Errorf("Error getting apiserver addresses of cluster: %s", err.Error())
			// sleep a while to try again, avoid trying in loop
			time.Sleep(30 * time.Second)
			continue
		}
		blog.Infof("apiserver addresses: %s", serverAddresses)

		bkeURL := getBkeAgentInfo()
		blog.Infof("bke-server url：%s", bkeURL)

		clusterInfoParams := ClusterInfoParams{
			ServerKey:     clusterID,
			ClusterID:     clusterID,
			ClientModule:  kubeAgentModule,
			ServerAddress: serverAddresses,
			CaCertData:    string(cfg.CAData),
			UserToken:     cfg.BearerToken,
		}

		var request *gorequest.SuperAgent
		insecureSkipVerify := viper.GetBool("agent.insecureSkipVerify")
		if insecureSkipVerify {
			request = gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: insecureSkipVerify})
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
				break
			}
			request = gorequest.New().TLSClientConfig(tlsConfig)
		}
		userToken := os.Getenv("USER_TOKEN")
		if len(userToken) != 0 {
			request.AppendHeader("Authorization", "Bearer "+userToken)
		}

		resp, respBody, errs := request.Put(bkeURL).Send(clusterInfoParams).End()
		if len(errs) > 0 {
			blog.Errorf("unable to connect to the bke server: %s", errs[0].Error())
			// sleep a while to try again, avoid trying in loop
			time.Sleep(30 * time.Second)
			continue
		}
		if resp.StatusCode >= 400 {
			codeName := json.Get([]byte(respBody), "code_name").ToString()
			message := json.Get([]byte(respBody), "message").ToString()
			blog.Errorf("Error updating cluster credential to bke, response code: %s, response message: %s", codeName, message)
		}

		select {
		case <-monitorTicker.C:
		}
	}
}

// get the k8s cluster apiserver addresses
func getApiserverAdresses(kubeClient *kubernetes.Clientset) (string, error) {
	var apiserverPort int32
	var endpointsList []string
	var serverAddresses string

	externalProxyAddresses := viper.GetString("agent.external-proxy-addresses")
	if externalProxyAddresses == "" {
		endpoints, err := kubeClient.CoreV1().Endpoints(defaultNamespace).Get(
			context.TODO(), clusterServiceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		for _, subset := range endpoints.Subsets {
			if len(subset.Addresses) == 0 {
				continue
			}

			// here we only use the apiserver secure-port
			for _, port := range subset.Ports {
				if port.Name == "https" {
					apiserverPort = port.Port
					break
				}
			}

			for _, addr := range subset.Addresses {
				err := pingEndpoint(net.JoinHostPort(addr.IP, strconv.Itoa(int(apiserverPort))))
				if err == nil {
					endpoint := "https://" + net.JoinHostPort(addr.IP, strconv.Itoa(int(apiserverPort)))
					endpointsList = append(endpointsList, endpoint)
				}
			}
		}
		sort.Strings(endpointsList)
		serverAddresses = strings.Join(endpointsList, ",")
	} else {
		serverSlice := strings.Split(externalProxyAddresses, ",")
		for _, server := range serverSlice {
			if !strings.HasPrefix(server, "https://") {
				return "", fmt.Errorf("got invalid external-proxy-addresses")
			}
		}
		serverAddresses = externalProxyAddresses
	}

	return serverAddresses, nil
}

func getBkeAgentInfo() string {
	bkeServerAddress := viper.GetString("bke.serverAddress")
	bkeReportPath := viper.GetString("bke.report-path")
	bkeURL := fmt.Sprintf("%s"+bkeReportPath, bkeServerAddress)

	return bkeURL
}

// probe the health of the apiserver address for 3 times
func pingEndpoint(host string) error {
	var err error
	for i := 0; i < 3; i++ {
		err = dialTLS(host)
		if err != nil && strings.Contains(err.Error(), "connection refused") {
			blog.Infof("Error connecting the apiserver %s. Retrying...: %s", host, err.Error())
			time.Sleep(time.Second)
			continue
		} else if err != nil {
			blog.Errorf("Error connecting the apiserver %s: %s", host, err.Error())
			return err
		}
		return err
	}
	return err
}

func dialTLS(host string) error {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", host, conf)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}
