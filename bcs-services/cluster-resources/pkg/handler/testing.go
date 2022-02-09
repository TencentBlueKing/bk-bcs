/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package handler

import (
	spb "google.golang.org/protobuf/types/known/structpb"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func genResListReq() clusterRes.ResListReq {
	return clusterRes.ResListReq{
		ProjectID: util.GetTestProjectID(),
		ClusterID: util.GetTestClusterID(),
		Namespace: util.GetTestNamespace(),
	}
}

func genResCreateReq(manifest *spb.Struct) clusterRes.ResCreateReq {
	return clusterRes.ResCreateReq{
		ProjectID: util.GetTestProjectID(),
		ClusterID: util.GetTestClusterID(),
		Manifest:  manifest,
	}
}

func genResUpdateReq(manifest *spb.Struct, name string) clusterRes.ResUpdateReq {
	return clusterRes.ResUpdateReq{
		ProjectID: util.GetTestProjectID(),
		ClusterID: util.GetTestClusterID(),
		Namespace: util.GetTestNamespace(),
		Name:      name,
		Manifest:  manifest,
	}
}

func genResGetReq(name string) clusterRes.ResGetReq {
	return clusterRes.ResGetReq{
		ProjectID: util.GetTestProjectID(),
		ClusterID: util.GetTestClusterID(),
		Namespace: util.GetTestNamespace(),
		Name:      name,
	}
}

func genResDeleteReq(name string) clusterRes.ResDeleteReq {
	return clusterRes.ResDeleteReq{
		ProjectID: util.GetTestProjectID(),
		ClusterID: util.GetTestClusterID(),
		Namespace: util.GetTestNamespace(),
		Name:      name,
	}
}
