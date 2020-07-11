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

package v1

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/utils"
)

type Storage interface {
	ListApplication(clusterID string, condition url.Values) (ApplicationList, error)
	ListProcess(clusterID string, condition url.Values) (ProcessList, error)
	ListTaskGroup(clusterID string, condition url.Values) (TaskGroupList, error)
	ListConfigMap(clusterID string, condition url.Values) (ConfigMapList, error)
	ListSecret(clusterID string, condition url.Values) (SecretList, error)
	ListService(clusterID string, condition url.Values) (ServiceList, error)
	ListEndpoint(clusterID string, condition url.Values) (EndpointList, error)
	ListDeployment(clusterID string, condition url.Values) (DeploymentList, error)
	ListNamespace(clusterID string, condition url.Values) ([]string, error)
	ListIPPoolStatic(clusterID string, condition url.Values) (IPPoolStaticList, error)
	ListIPPoolStaticDetail(clusterID string, condition url.Values) (IPPoolStaticDetailList, error)

	InspectApplication(clusterID, namespace, name string) (*ApplicationSet, error)
	InspectProcess(clusterID, namespace, name string) (*ProcessSet, error)
	InspectTaskGroup(clusterID, namespace, name string) (*TaskGroupSet, error)
	InspectConfigMap(clusterID, namespace, name string) (*ConfigMapSet, error)
	InspectSecret(clusterID, namespace, name string) (*SecretSet, error)
	InspectService(clusterID, namespace, name string) (*ServiceSet, error)
	InspectEndpoint(clusterID, namespace, name string) (*EndpointSet, error)
	InspectDeployment(clusterID, namespace, name string) (*DeploymentSet, error)
}

const (
	BcsStorageListDynamicURI    = "%s/bcsapi/v4/storage/query/mesos/dynamic/clusters/%s/%s"
	BcsStorageInspectDynamicURI = "%s/bcsapi/v4/storage/mesos/dynamic/namespace_resources/clusters/%s/namespaces/%s/%s/%s"
)

type bcsStorage struct {
	bcsApiAddress string
	requester     utils.ApiRequester
}

func NewBcsStorage(options types.ClientOptions) Storage {
	return &bcsStorage{
		bcsApiAddress: options.BcsApiAddress,
		requester:     utils.NewApiRequester(options.ClientSSL, options.BcsToken),
	}
}

func (bs *bcsStorage) ListApplication(clusterID string, condition url.Values) (ApplicationList, error) {
	data, err := bs.listResource(clusterID, BcsStorageDynamicTypeApplication, condition)
	if err != nil {
		return nil, err
	}

	var result ApplicationList
	err = codec.DecJson(data, &result)
	return result, err
}

func (bs *bcsStorage) ListProcess(clusterID string, condition url.Values) (ProcessList, error) {
	data, err := bs.listResource(clusterID, BcsStorageDynamicTypeProcess, condition)
	if err != nil {
		return nil, err
	}

	var result ProcessList
	err = codec.DecJson(data, &result)
	return result, err
}

func (bs *bcsStorage) ListTaskGroup(clusterID string, condition url.Values) (TaskGroupList, error) {
	data, err := bs.listResource(clusterID, BcsStorageDynamicTypeTaskGroup, condition)
	if err != nil {
		return nil, err
	}

	var result TaskGroupList
	err = codec.DecJson(data, &result)
	return result, err
}

func (bs *bcsStorage) ListConfigMap(clusterID string, condition url.Values) (ConfigMapList, error) {
	data, err := bs.listResource(clusterID, BcsStorageDynamicTypeConfigMap, condition)
	if err != nil {
		return nil, err
	}

	var result ConfigMapList
	err = codec.DecJson(data, &result)
	return result, err
}

func (bs *bcsStorage) ListSecret(clusterID string, condition url.Values) (SecretList, error) {
	data, err := bs.listResource(clusterID, BcsStorageDynamicTypeSecret, condition)
	if err != nil {
		return nil, err
	}

	var result SecretList
	err = codec.DecJson(data, &result)
	return result, err
}

func (bs *bcsStorage) ListService(clusterID string, condition url.Values) (ServiceList, error) {
	data, err := bs.listResource(clusterID, BcsStorageDynamicTypeService, condition)
	if err != nil {
		return nil, err
	}

	var result ServiceList
	err = codec.DecJson(data, &result)
	return result, err
}

func (bs *bcsStorage) ListEndpoint(clusterID string, condition url.Values) (EndpointList, error) {
	data, err := bs.listResource(clusterID, BcsStorageDynamicTypeEndpoint, condition)
	if err != nil {
		return nil, err
	}

	var result EndpointList
	err = codec.DecJson(data, &result)
	return result, err
}

func (bs *bcsStorage) ListDeployment(clusterID string, condition url.Values) (DeploymentList, error) {
	data, err := bs.listResource(clusterID, BcsStorageDynamicTypeDeployment, condition)
	if err != nil {
		return nil, err
	}

	var result DeploymentList
	err = codec.DecJson(data, &result)
	return result, err
}

func (bs *bcsStorage) ListNamespace(clusterID string, condition url.Values) ([]string, error) {
	data, err := bs.listResource(clusterID, BcsStorageDynamicTypeNamespace, condition)
	if err != nil {
		return nil, err
	}

	var result []string
	err = codec.DecJson(data, &result)
	return result, err
}

// ListIPPoolStatic query netservice ip pool static resource data from storage.
func (bs *bcsStorage) ListIPPoolStatic(clusterID string, condition url.Values) (IPPoolStaticList, error) {
	data, err := bs.listResource(clusterID, BcsStorageDynamicTypeIPPoolStatic, condition)
	if err != nil {
		return nil, err
	}

	var result IPPoolStaticList
	err = codec.DecJson(data, &result)
	return result, err
}

// ListIPPoolStaticDetail query netservice ip pool static resource detail data from storage.
func (bs *bcsStorage) ListIPPoolStaticDetail(clusterID string, condition url.Values) (IPPoolStaticDetailList, error) {
	data, err := bs.listResource(clusterID, BcsStorageDynamicTypeIPPoolStaticDetail, condition)
	if err != nil {
		return nil, err
	}

	var result IPPoolStaticDetailList
	err = codec.DecJson(data, &result)
	return result, err
}

func (bs *bcsStorage) InspectApplication(clusterID, namespace, name string) (*ApplicationSet, error) {
	data, err := bs.inspectResource(clusterID, namespace, BcsStorageDynamicTypeApplication, name)
	if err != nil {
		return nil, err
	}

	var result ApplicationSet
	err = codec.DecJson(data, &result)
	return &result, err
}

func (bs *bcsStorage) InspectProcess(clusterID, namespace, name string) (*ProcessSet, error) {
	data, err := bs.inspectResource(clusterID, namespace, BcsStorageDynamicTypeProcess, name)
	if err != nil {
		return nil, err
	}

	var result ProcessSet
	err = codec.DecJson(data, &result)
	return &result, err
}

func (bs *bcsStorage) InspectTaskGroup(clusterID, namespace, name string) (*TaskGroupSet, error) {
	data, err := bs.inspectResource(clusterID, namespace, BcsStorageDynamicTypeTaskGroup, name)
	if err != nil {
		return nil, err
	}

	var result TaskGroupSet
	err = codec.DecJson(data, &result)
	return &result, err
}

func (bs *bcsStorage) InspectConfigMap(clusterID, namespace, name string) (*ConfigMapSet, error) {
	data, err := bs.inspectResource(clusterID, namespace, BcsStorageDynamicTypeConfigMap, name)
	if err != nil {
		return nil, err
	}

	var result ConfigMapSet
	err = codec.DecJson(data, &result)
	return &result, err
}

func (bs *bcsStorage) InspectSecret(clusterID, namespace, name string) (*SecretSet, error) {
	data, err := bs.inspectResource(clusterID, namespace, BcsStorageDynamicTypeSecret, name)
	if err != nil {
		return nil, err
	}

	var result SecretSet
	err = codec.DecJson(data, &result)
	return &result, err
}

func (bs *bcsStorage) InspectService(clusterID, namespace, name string) (*ServiceSet, error) {
	data, err := bs.inspectResource(clusterID, namespace, BcsStorageDynamicTypeService, name)
	if err != nil {
		return nil, err
	}

	var result ServiceSet
	err = codec.DecJson(data, &result)
	return &result, err
}

func (bs *bcsStorage) InspectEndpoint(clusterID, namespace, name string) (*EndpointSet, error) {
	data, err := bs.inspectResource(clusterID, namespace, BcsStorageDynamicTypeEndpoint, name)
	if err != nil {
		return nil, err
	}

	var result EndpointSet
	err = codec.DecJson(data, &result)
	return &result, err
}

func (bs *bcsStorage) InspectDeployment(clusterID, namespace, name string) (*DeploymentSet, error) {
	data, err := bs.inspectResource(clusterID, namespace, BcsStorageDynamicTypeDeployment, name)
	if err != nil {
		return nil, err
	}

	var result DeploymentSet
	err = codec.DecJson(data, &result)
	return &result, err
}

func (bs *bcsStorage) listResource(clusterID, resourceType string, condition url.Values) ([]byte, error) {
	if condition == nil {
		condition = make(url.Values)
	}

	conditionMap := make(map[string]string)
	for k, v := range condition {
		conditionMap[k] = strings.Join(v, ",")
	}

	var data []byte
	if err := codec.EncJson(conditionMap, &data); err != nil {
		return nil, err
	}

	// namespace not need to post
	method := http.MethodPost
	if resourceType == BcsStorageDynamicTypeNamespace {
		method = http.MethodGet
	}
	resp, err := bs.requester.Do(
		fmt.Sprintf(BcsStorageListDynamicURI, bs.bcsApiAddress, clusterID, resourceType),
		method,
		data,
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
		return nil, fmt.Errorf("list dynamic %s failed: %s", resourceType, msg)
	}

	return data, nil
}

func (bs *bcsStorage) inspectResource(clusterID, namespace, resourceType, name string) ([]byte, error) {
	resp, err := bs.requester.Do(
		fmt.Sprintf(BcsStorageInspectDynamicURI, bs.bcsApiAddress, clusterID, namespace, resourceType, name),
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
		return nil, fmt.Errorf("inspect dynamic %s failed: %s", resourceType, msg)
	}

	return data, nil
}
