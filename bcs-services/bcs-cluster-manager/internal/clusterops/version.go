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

package clusterops

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckClusterConnection check cluster connection by version
func (ko *K8SOperator) CheckClusterConnection(ctx context.Context, clusterID string) error {
	if ko == nil {
		return ErrServerNotInit
	}
	clientInterface, err := ko.GetClusterClient(clusterID)
	if err != nil {
		blog.Errorf("CheckClusterConnection[%s] GetClusterClient failed: %v", clusterID, err)
		return err
	}

	_, err = clientInterface.CoreV1().Namespaces().List(ctx, v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("CheckClusterConnection[%s] failed: %v", clusterID, err)
	}

	blog.Infof("CheckClusterConnection[%s] success", clusterID)
	return nil
}
