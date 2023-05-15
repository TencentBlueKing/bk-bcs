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

package k8sclient

import (
	"context"

	logv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/apis/bkbcs/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListBcsLogConfig list bcslogconfigs
func ListBcsLogConfig(ctx context.Context, clusterID string) ([]logv1.BcsLogConfig, error) {
	client, err := GetKubebkbcsClientByClusterID(clusterID)
	if err != nil {
		return nil, err
	}
	list, err := client.BkbcsV1().BcsLogConfigs("").List(ctx, v1.ListOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return list.Items, nil
}

// GetBcsLogConfig get bcslogconfigs
func GetBcsLogConfig(ctx context.Context, clusterID, namespace, name string) (*logv1.BcsLogConfig, error) {
	client, err := GetKubebkbcsClientByClusterID(clusterID)
	if err != nil {
		return nil, err
	}
	blc, err := client.BkbcsV1().BcsLogConfigs(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return blc, nil
}

// DeleteBcsLogConfig delete bcslogconfigs
func DeleteBcsLogConfig(ctx context.Context, clusterID, namespace, name string) error {
	client, err := GetKubebkbcsClientByClusterID(clusterID)
	if err != nil {
		return err
	}
	return client.BkbcsV1().BcsLogConfigs(namespace).Delete(ctx, name, v1.DeleteOptions{})
}
