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

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	commonTypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
)

func (bs *bcsScheduler) GetApplicationDefinition(clusterID, namespace, name string) (*commonTypes.ReplicaController, error) {
	return bs.getApplicationDefinition(clusterID, namespace, name)
}

func (bs *bcsScheduler) GetProcessDefinition(clusterID, namespace, name string) (*commonTypes.ReplicaController, error) {
	return bs.getProcessDefinition(clusterID, namespace, name)
}

func (bs *bcsScheduler) GetDeploymentDefinition(clusterID, namespace, name string) (*commonTypes.BcsDeployment, error) {
	return bs.getDeploymentDefinition(clusterID, namespace, name)
}

func (bs *bcsScheduler) getApplicationDefinition(clusterID, namespace, name string) (*commonTypes.ReplicaController, error) {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerAppDefinitionURI, bs.bcsAPIAddress, namespace, name),
		http.MethodGet,
		nil,
		getClusterIDHeader(clusterID),
	)

	if err != nil {
		return nil, err
	}

	code, msg, data, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	if code != 0 {
		return nil, fmt.Errorf("get application definition failed: %s", msg)
	}

	var result commonTypes.ReplicaController
	err = codec.DecJson(data, &result)

	if result.Kind != "" && result.Kind != commonTypes.BcsDataType_APP {
		return nil, fmt.Errorf("there is no such application")
	}
	return &result, err
}

func (bs *bcsScheduler) getProcessDefinition(clusterID, namespace, name string) (*commonTypes.ReplicaController, error) {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerAppDefinitionURI, bs.bcsAPIAddress, namespace, name),
		http.MethodGet,
		nil,
		getClusterIDHeader(clusterID),
	)

	if err != nil {
		return nil, err
	}

	code, msg, data, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	if code != 0 {
		return nil, fmt.Errorf("get process definition failed: %s", msg)
	}

	var result commonTypes.ReplicaController
	err = codec.DecJson(data, &result)

	if result.Kind != commonTypes.BcsDataType_PROCESS {
		return nil, fmt.Errorf("there is no such process")
	}
	return &result, err
}

func (bs *bcsScheduler) getDeploymentDefinition(clusterID, namespace, name string) (*commonTypes.BcsDeployment, error) {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerDeployDefinitionURI, bs.bcsAPIAddress, namespace, name),
		http.MethodGet,
		nil,
		getClusterIDHeader(clusterID),
	)

	if err != nil {
		return nil, err
	}

	code, msg, data, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	if code != 0 {
		return nil, fmt.Errorf("get deployment definition failed: %s", msg)
	}

	var result commonTypes.BcsDeployment
	err = codec.DecJson(data, &result)
	return &result, err
}
