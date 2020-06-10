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

func (bs *bcsScheduler) DeleteApplication(clusterID, namespace, name string, enforce bool) error {
	return bs.deleteResource(clusterID, namespace, BcsSchedulerResourceApplication, name, enforce)
}

func (bs *bcsScheduler) DeleteProcess(clusterID, namespace, name string, enforce bool) error {
	return bs.deleteResource(clusterID, namespace, BcsSchedulerResourceProcess, name, enforce)
}

func (bs *bcsScheduler) DeleteConfigMap(clusterID, namespace, name string, enforce bool) error {
	return bs.deleteResource(clusterID, namespace, BcsSchedulerResourceConfigMap, name, enforce)
}

func (bs *bcsScheduler) DeleteSecret(clusterID, namespace, name string, enforce bool) error {
	return bs.deleteResource(clusterID, namespace, BcsSchedulerResourceSecret, name, enforce)
}

func (bs *bcsScheduler) DeleteService(clusterID, namespace, name string, enforce bool) error {
	return bs.deleteResource(clusterID, namespace, BcsSchedulerResourceService, name, enforce)
}

func (bs *bcsScheduler) DeleteDeployment(clusterID, namespace, name string, enforce bool) error {
	return bs.deleteResource(clusterID, namespace, BcsSchedulerResourceDeployment, name, enforce)
}

func (bs *bcsScheduler) DeleteDaemonset(clusterID, namespace, name string, enforce bool) error {
	return bs.deleteResource(clusterID, namespace, BcsSchedulerResourceDaemonset, name, enforce)
}

func (bs *bcsScheduler) deleteResource(clusterID, namespace, resourceType, name string, enforce bool) error {
	enforceNum := 0
	if enforce {
		enforceNum = 1
	}

	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerDeleteResourceURI, bs.bcsAPIAddress, namespace, resourceType, name, enforceNum),
		http.MethodDelete,
		nil,
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
		return fmt.Errorf("delete resource %s failed: %s", resourceType, msg)
	}

	return nil
}
