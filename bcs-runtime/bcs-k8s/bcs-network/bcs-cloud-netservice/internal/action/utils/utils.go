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

// Package utils is tool functions for action
package utils

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store/kube"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
)

func listClusterIPObject(ctx context.Context, storeIf store.Interface, labels map[string]string) (
	[]*types.IPObject, error) {
	ipList, err := storeIf.ListIPObject(ctx, labels)
	if err != nil {
		blog.Errorf("failed to list ips by %v, err %s", labels, err.Error())
		return nil, fmt.Errorf("failed to list ips by %v, err %s", labels, err.Error())
	}
	return ipList, nil
}

// CheckIPQuota returned boolean indicates whether the cluster has enough ip quota
func CheckIPQuota(ctx context.Context, storeIf store.Interface, cluster string) error {
	var err error
	var quota *types.IPQuota
	var activeNonFixedIPs []*types.IPObject
	var fixedIPs []*types.IPObject
	var eniIPs []*types.IPObject
	quota, err = storeIf.GetIPQuota(ctx, cluster)
	if err != nil {
		return err
	}
	activeNonFixedIPs, err = listClusterIPObject(ctx, storeIf, map[string]string{
		kube.CrdNameLabelsCluster: cluster,
		kube.CrdNameLabelsStatus:  types.IPStatusActive,
		kube.CrdNameLabelsIsFixed: strconv.FormatBool(false),
	})
	if err != nil {
		return err
	}
	fixedIPs, err = listClusterIPObject(ctx, storeIf, map[string]string{
		kube.CrdNameLabelsCluster: cluster,
		kube.CrdNameLabelsIsFixed: strconv.FormatBool(true),
	})
	if err != nil {
		return err
	}
	eniIPs, err = listClusterIPObject(ctx, storeIf, map[string]string{
		kube.CrdNameLabelsCluster: cluster,
		kube.CrdNameLabelsStatus:  types.IPStatusENIPrimary,
	})
	if err != nil {
		return err
	}
	if len(activeNonFixedIPs)+len(fixedIPs)+len(eniIPs) < int(quota.Limit) {
		return nil
	}
	return fmt.Errorf("cluster %s quota %d is all used, active-non-fixed-ip %d,  fixed-ip %d, eni-ip %d",
		cluster, quota.Limit, len(activeNonFixedIPs), len(fixedIPs), len(eniIPs))
}
