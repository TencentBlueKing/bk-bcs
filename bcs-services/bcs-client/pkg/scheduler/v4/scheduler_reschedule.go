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

func (bs *bcsScheduler) RescheduleTaskGroup(clusterID, namespace, applicationName, taskGroupName string) error {
	return bs.rescheduleTaskGroup(clusterID, namespace, applicationName, taskGroupName)
}

func (bs *bcsScheduler) rescheduleTaskGroup(clusterID, namespace, applicationName, taskGroupName string) error {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerRescheduleURI, bs.bcsAPIAddress, namespace, applicationName, taskGroupName),
		http.MethodPut,
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
		return fmt.Errorf("reschedule taskgroup failed: %s", msg)
	}

	return nil
}
