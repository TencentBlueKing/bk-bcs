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
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	tokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

// Run run agent
func Run() error {
	kubeconfig := viper.GetString("agent.kubeconfig")
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("error getting k8s cluster config: %s", err.Error())
	}

	if kubeconfig == "" {
		// since go-client 9.0.0, the restclient.Config returned by BuildConfigFromFlags doesn't have BearerToken, so manually get the BearerToken
		token, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			return fmt.Errorf("error getting the BearerToken: %s", err.Error())
		}
		cfg.BearerToken = string(token)
		if err := populateCAData(cfg); err != nil {
			return fmt.Errorf("error populating ca data: %s", err.Error())
		}
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building kubernetes clientset: %s", err.Error())
	}

	useWebsocket := viper.GetBool("agent.use-websocket")
	if useWebsocket {
		err := buildWebsocketToBke(cfg)
		if err != nil {
			return err
		}
	} else {
		go reportToBke(kubeClient, cfg)
	}

	http.Handle("/metrics", promhttp.Handler())
	listenAddr := viper.GetString("agent.listenAddr")
	return http.ListenAndServe(listenAddr, nil)
}

func populateCAData(cfg *rest.Config) error {
	bytes, err := ioutil.ReadFile(cfg.CAFile)
	if err != nil {
		return err
	}
	cfg.CAData = bytes
	return nil
}
