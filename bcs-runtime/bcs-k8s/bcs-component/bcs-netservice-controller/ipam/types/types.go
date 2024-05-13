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

// Package types is types for ipam
package types

import (
	"net"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
)

// K8sArgs args for k8s
type K8sArgs struct {
	types.CommonArgs
	IP                         net.IP
	K8S_POD_NAME               types.UnmarshallableString
	K8S_POD_NAMESPACE          types.UnmarshallableString
	K8S_POD_INFRA_CONTAINER_ID types.UnmarshallableString
}

// LoadK8sArgs gets k8s related args from CNI args
func LoadK8sArgs(args *skel.CmdArgs) (*K8sArgs, error) {
	k8sArgs := &K8sArgs{}

	blog.Infof("get K8sArgs: %v", args)
	err := types.LoadArgs(args.Args, k8sArgs)
	if err != nil {
		return nil, err
	}

	return k8sArgs, nil
}
