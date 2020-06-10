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

package v4

import (
	"fmt"
	"net/http"
)

func (bs *bcsScheduler) CreateApplication(clusterID, namespace string, data []byte) error {
	return bs.createResource(clusterID, namespace, BcsSchedulerResourceApplication, data)
}

func (bs *bcsScheduler) CreateProcess(clusterID, namespace string, data []byte) error {
	return bs.createResource(clusterID, namespace, BcsSchedulerResourceProcess, data)
}

func (bs *bcsScheduler) CreateConfigMap(clusterID, namespace string, data []byte) error {
	return bs.createResource(clusterID, namespace, BcsSchedulerResourceConfigMap, data)
}

func (bs *bcsScheduler) CreateSecret(clusterID, namespace string, data []byte) error {
	return bs.createResource(clusterID, namespace, BcsSchedulerResourceSecret, data)
}

func (bs *bcsScheduler) CreateService(clusterID, namespace string, data []byte) error {
	return bs.createResource(clusterID, namespace, BcsSchedulerResourceService, data)
}

func (bs *bcsScheduler) CreateDeployment(clusterID, namespace string, data []byte) error {
	return bs.createResource(clusterID, namespace, BcsSchedulerResourceDeployment, data)
}

func (bs *bcsScheduler) CreateDaemonset(clusterID, namespace string, data []byte) error {
	return bs.createResource(clusterID, namespace, BcsSchedulerResourceDaemonset, data)
}

func (bs *bcsScheduler) createResource(clusterID, namespace, resourceType string, data []byte) error {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerResourceURI, bs.bcsAPIAddress, namespace, resourceType, ""),
		http.MethodPost,
		data,
		getClusterIDHeader(clusterID),
	)

	if err != nil {
		return err
	}

	code, msg, _, err := parseResponse(resp)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("create resource %s failed: %s", resourceType, msg)
	}

	return nil
}
