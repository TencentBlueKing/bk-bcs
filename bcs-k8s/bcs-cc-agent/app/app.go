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
	"os"
	"os/signal"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-cc-agent/config"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-cc-agent/options"
)

//Run bcs-cc-agent server
func Run(op *options.ServerOption) {
	conf, err := parseConfig(op)
	if err != nil {
		fmt.Printf("parse config failed: %v\n", err)
		os.Exit(1)
	}

	if conf.EngineType == "kubernetes" {
		err := synchronizeK8sNodeInfo(conf)
		if err != nil {
			fmt.Printf("synchronize failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		err := synchronizeMesosNodeInfo(conf)
		if err != nil {
			fmt.Printf("synchronize failed: %v\n", err)
			os.Exit(1)
		}
	}

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	blog.Infof("Got OS shutdown signal, shutting down bcs-cc-agent server gracefully...")

	return
}

func parseConfig(op *options.ServerOption) (*config.BcsCcAgentConfig, error) {
	bcsCcAgentConfig := config.NewBcsCcAgentConfig()
	bcsCcAgentConfig.MetricPort = op.MetricPort
	bcsCcAgentConfig.EngineType = op.EngineType
	bcsCcAgentConfig.Kubeconfig = op.KubeConfig
	bcsCcAgentConfig.KubeMaster = op.KubeMaster

	bcsCcAgentConfig.EsbUrl = op.EsbUrl
	appCode, err := encrypt.DesDecryptFromBase([]byte(op.AppCode))
	if err != nil {
		return nil, fmt.Errorf("unable to desdecrypt app_code [%s]: %s", op.AppCode, err.Error())
	}
	bcsCcAgentConfig.AppCode = string(appCode)
	appSecret, err := encrypt.DesDecryptFromBase([]byte(op.AppSecret))
	if err != nil {
		return nil, fmt.Errorf("unable to desdecrypt app_secret [%s]: %s", op.AppSecret, err.Error())
	}
	bcsCcAgentConfig.AppSecret = string(appSecret)
	bkUsername, err := encrypt.DesDecryptFromBase([]byte(op.BkUsername))
	if err != nil {
		return nil, fmt.Errorf("unable to desdecrypt bk username [%s]: %s", op.BkUsername, err.Error())
	}
	bcsCcAgentConfig.BkUsername = string(bkUsername)

	return bcsCcAgentConfig, nil
}
