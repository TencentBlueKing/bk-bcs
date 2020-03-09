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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"bk-bcs/bcs-common/common/blog"
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
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type ClusterInfoParams struct {
	RegisterToken   string `json:"register_token"`
	ServerAddresses string `json:"server_addresses"`
	CaCertData      string `json:"cacert_data"`
	UserToken       string `json:"user_token"`
}

func reportToBke(kubeClient *kubernetes.Clientset, cfg *rest.Config) {
	periodSync := viper.GetInt("agent.periodSync")
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

		bkeUrl, registerToken := getBkeAgentInfo()
		blog.Infof("bke-server url：%s", bkeUrl)

		clusterInfoParams := ClusterInfoParams{
			RegisterToken:   registerToken,
			ServerAddresses: serverAddresses,
			CaCertData:      string(cfg.CAData),
			UserToken:       cfg.BearerToken,
		}

		var request *gorequest.SuperAgent
		insecureSkipVerify := viper.GetBool("agent.insecureSkipVerify")
		if insecureSkipVerify {
			request = gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: insecureSkipVerify})
		} else {
			pool := x509.NewCertPool()
			caCrtStr := os.Getenv("SERVER_CERT")
			caCrt := []byte(caCrtStr)
			pool.AppendCertsFromPEM(caCrt)
			request = gorequest.New().TLSClientConfig(&tls.Config{RootCAs: pool})
		}

		resp, respBody, errs := request.Put(bkeUrl).Send(clusterInfoParams).End()
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
		endpoints, err := kubeClient.CoreV1().Endpoints(defaultNamespace).Get(clusterServiceName, metav1.GetOptions{})
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
		serverAddresses = strings.Join(endpointsList, ";")
	} else {
		serverSlice := strings.Split(externalProxyAddresses, ";")
		for _, server := range serverSlice {
			if !strings.HasPrefix(server, "https://") {
				return "", fmt.Errorf("got invalid external-proxy-addresses")
			}
		}
		serverAddresses = externalProxyAddresses
	}

	return serverAddresses, nil
}

func getBkeAgentInfo() (string, string) {
	bkeServerAddress := viper.GetString("bke.serverAddress")
	clusterId := viper.GetString("cluster.id")
	registerToken := os.Getenv("REGISTER_TOKEN")

	bkeUrl := fmt.Sprintf("%s/rest/clusters/%s/credentials", bkeServerAddress, clusterId)

	return bkeUrl, registerToken
}

// probe the health of the apiserver address for 3 times
func pingEndpoint(host string) error {
	var err error
	for i := 0; i < 3; i++ {
		err = dialTls(host)
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

func dialTls(host string) error {
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
