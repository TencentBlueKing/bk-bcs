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

func (bs *bcsScheduler) ResumeDeployment(clusterID, namespace, name string) error {
	return bs.resumeDeployment(clusterID, namespace, name)
}

func (bs *bcsScheduler) CancelDeployment(clusterID, namespace, name string) error {
	return bs.cancelDeployment(clusterID, namespace, name)
}

func (bs *bcsScheduler) PauseDeployment(clusterID, namespace, name string) error {
	return bs.pauseDeployment(clusterID, namespace, name)
}

func (bs *bcsScheduler) resumeDeployment(clusterID, namespace, name string) error {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerResumeDeploymentURI, bs.bcsAPIAddress, namespace, name),
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
		return fmt.Errorf("resume deployment failed: %s", msg)
	}

	return nil
}

func (bs *bcsScheduler) cancelDeployment(clusterID, namespace, name string) error {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerCancelDeploymentURI, bs.bcsAPIAddress, namespace, name),
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
		return fmt.Errorf("cancel deployment failed: %s", msg)
	}

	return nil
}

func (bs *bcsScheduler) pauseDeployment(clusterID, namespace, name string) error {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerPauseDeploymentURI, bs.bcsAPIAddress, namespace, name),
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
		return fmt.Errorf("pause deployment failed: %s", msg)
	}

	return nil
}
