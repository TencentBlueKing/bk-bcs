/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/sdk"
)

func main() {
	var proxyAddr, serverAddr, clusterID string

	flag.StringVar(&proxyAddr, "proxy_address", "", "address of proxy")
	flag.StringVar(&serverAddr, "server_address", "", "address of server")
	flag.StringVar(&clusterID, "clusterid", "", "clusterid of the cluster under server's management")
	flag.Parse()

	client := sdk.NewWebsocketClient(proxyAddr, serverAddr, clusterID)
	client.Start()
	for {
		select {}
	}
}
