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
	"net/url"
)

func (bs *bcsScheduler) UpdateApplication(clusterID, namespace string, data []byte, extraValue url.Values) error {
	return bs.updateResource(clusterID, namespace, BcsSchedulerResourceApplication, data, extraValue)
}

func (bs *bcsScheduler) UpdateProcess(clusterID, namespace string, data []byte, extraValue url.Values) error {
	return bs.updateResource(clusterID, namespace, BcsSchedulerResourceProcess, data, extraValue)
}

func (bs *bcsScheduler) UpdateConfigMap(clusterID, namespace string, data []byte, extraValue url.Values) error {
	return bs.updateResource(clusterID, namespace, BcsSchedulerResourceConfigMap, data, extraValue)
}

func (bs *bcsScheduler) UpdateSecret(clusterID, namespace string, data []byte, extraValue url.Values) error {
	return bs.updateResource(clusterID, namespace, BcsSchedulerResourceSecret, data, extraValue)
}

func (bs *bcsScheduler) UpdateService(clusterID, namespace string, data []byte, extraValue url.Values) error {
	return bs.updateResource(clusterID, namespace, BcsSchedulerResourceService, data, extraValue)
}

func (bs *bcsScheduler) UpdateDeployment(clusterID, namespace string, data []byte, extraValue url.Values) error {
	return bs.updateResource(clusterID, namespace, BcsSchedulerResourceDeployment, data, extraValue)
}

func (bs *bcsScheduler) updateResource(clusterID, namespace, resourceType string, data []byte, extraValue url.Values) error {
	if extraValue == nil {
		extraValue = make(url.Values)
	}
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerResourceURI, bs.bcsAPIAddress, namespace, resourceType, extraValue.Encode()),
		http.MethodPut,
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
		return fmt.Errorf("update resource %s failed: %s", resourceType, msg)
	}

	return nil
}
