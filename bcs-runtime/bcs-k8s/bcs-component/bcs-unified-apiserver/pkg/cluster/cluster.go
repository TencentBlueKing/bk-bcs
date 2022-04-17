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

package cluster

import (
	"errors"
	"net/http"

	apirequest "k8s.io/apiserver/pkg/endpoints/request"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/cluster/isolated"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/config"
)

func ClusterFactory(clusterId string, reqInfo *apirequest.RequestInfo, uri string) (http.Handler, error) {
	cluster, ok := config.G.GetMember(clusterId)
	if !ok {
		return nil, errors.New("invalid cluster")
	}

	var (
		handle http.Handler
		err    error
	)

	switch cluster.Kind {
	case string(config.IsolatedCLuster):
		handle, err = isolated.NewHandler(cluster.Member)
	case string(config.SharedCluster):
		handle, err = isolated.NewHandler(cluster.Member)
	case string(config.FederatedCluter):
		handle, err = isolated.NewHandler(cluster.Member)
	default:
		return nil, errors.New("not valid cluster kind")
	}
	return handle, err

}
