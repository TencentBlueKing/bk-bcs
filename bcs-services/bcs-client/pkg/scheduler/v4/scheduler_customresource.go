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
	"encoding/json"
	"fmt"
	"net/http"

	simplejson "github.com/bitly/go-simplejson"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

//CreateResourceDefinition create CRD by definition file
func (bs *bcsScheduler) CreateCustomResourceDefinition(clusterID string, data []byte) error {
	resp, err := bs.requester.DoForResponse(
		fmt.Sprintf(bcsScheudlerCustomResourceDefinitionURL, bs.bcsAPIAddress),
		http.MethodPost,
		data,
		getClusterIDHeader(clusterID),
	)
	if err != nil {
		return err
	}
	//if there is no network error, bcs-api always response 200
	// we need to verify response body object
	// * if response is not json, low level http error
	// * if response object is Status, we can read Status.message
	// * Custom Resource Object when create successfully
	json, err := simplejson.NewJson(resp.Reply)
	if err != nil {
		return fmt.Errorf("response format is Not expected Json, http error: %s", string(resp.Reply))
	}
	replyKind, _ := json.Get("kind").String()
	if replyKind == StatusKind {
		msg, _ := json.Get("message").String()
		return fmt.Errorf("%s", msg)
	}
	//resply is CRD object, ommit now
	return nil
}

//UpdateResourceDefinition replace specified CRD
func (bs *bcsScheduler) UpdateCustomResourceDefinition(clusterID, name string, data []byte) error {
	if len(name) == 0 {
		return fmt.Errorf("Lost specified crd name")
	}
	//get exist data for version validation
	crd, err := bs.GetCustomResourceDefinition(clusterID, name)
	if err != nil {
		return err
	}
	udpateObject := &v1beta1.CustomResourceDefinition{}
	json.Unmarshal(data, udpateObject)
	udpateObject.ResourceVersion = crd.ResourceVersion
	updateData, _ := json.Marshal(udpateObject)
	//update work flow
	preURL := fmt.Sprintf(bcsScheudlerCustomResourceDefinitionURL, bs.bcsAPIAddress)
	reqURL := fmt.Sprintf("%s/%s", preURL, name)
	resp, err := bs.requester.DoForResponse(
		reqURL,
		http.MethodPut,
		updateData,
		getClusterIDHeader(clusterID),
	)
	if err != nil {
		return err
	}
	//if there is no network error, bcs-api always response 200
	// we need to verify response body object
	// * if response is not json, low level http error
	// * if response object is Status, we can read Status.message
	// * Custom Resource Object when create successfully
	json, err := simplejson.NewJson(resp.Reply)
	if err != nil {
		return fmt.Errorf("response format is Not expected Json, http error: %s", string(resp.Reply))
	}
	replyKind, _ := json.Get("kind").String()
	if replyKind == StatusKind {
		msg, _ := json.Get("message").String()
		return fmt.Errorf("%s", msg)
	}
	return nil
}

//ListCustomResourceDefinition list all created CRD
func (bs *bcsScheduler) ListCustomResourceDefinition(clusterID string) (*v1beta1.CustomResourceDefinitionList, error) {
	resp, err := bs.requester.DoForResponse(
		fmt.Sprintf(bcsScheudlerCustomResourceDefinitionURL, bs.bcsAPIAddress),
		http.MethodGet,
		nil,
		getClusterIDHeader(clusterID),
	)
	if err != nil {
		return nil, err
	}
	jsonObj, err := simplejson.NewJson(resp.Reply)
	if err != nil {
		return nil, fmt.Errorf("response format is Not expected Json, http error: %s", string(resp.Reply))
	}
	replyKind, _ := jsonObj.Get("kind").String()
	if replyKind == StatusKind {
		msg, _ := jsonObj.Get("message").String()
		return nil, fmt.Errorf("%s", msg)
	}
	crd := &v1beta1.CustomResourceDefinitionList{}
	json.Unmarshal(resp.Reply, crd)
	return crd, nil
}

//GetCustomResourceDefinition get specified CRD
func (bs *bcsScheduler) GetCustomResourceDefinition(clusterID string, name string) (*v1beta1.CustomResourceDefinition, error) {
	preURL := fmt.Sprintf(bcsScheudlerCustomResourceDefinitionURL, bs.bcsAPIAddress)
	reqURL := fmt.Sprintf("%s/%s", preURL, name)
	resp, err := bs.requester.DoForResponse(
		reqURL,
		http.MethodGet,
		nil,
		getClusterIDHeader(clusterID),
	)
	if err != nil {
		return nil, err
	}
	//if there is no network error, bcs-api always response 200
	// we need to verify response body object
	// * if response is not json, low level http error
	// * if response object is Status, we can read Status.message
	// * Custom Resource Object when create successfully
	jsonObj, err := simplejson.NewJson(resp.Reply)
	if err != nil {
		return nil, fmt.Errorf("response format is Not expected Json, http error: %s", string(resp.Reply))
	}
	replyKind, _ := jsonObj.Get("kind").String()
	if replyKind == StatusKind {
		msg, _ := jsonObj.Get("message").String()
		return nil, fmt.Errorf("%s", msg)
	}
	crd := &v1beta1.CustomResourceDefinition{}
	json.Unmarshal(resp.Reply, crd)
	//clean info
	crd.SelfLink = ""
	//crd.ResourceVersion = ""
	crd.Generation = 0
	crd.UID = ""
	return crd, nil
}

//DeleteCustomResourceDefinition delete specified CRD
func (bs *bcsScheduler) DeleteCustomResourceDefinition(clusterID, name string) error {
	preURL := fmt.Sprintf(bcsScheudlerCustomResourceDefinitionURL, bs.bcsAPIAddress)
	reqURL := fmt.Sprintf("%s/%s", preURL, name)
	resp, err := bs.requester.DoForResponse(
		reqURL,
		http.MethodDelete,
		nil,
		getClusterIDHeader(clusterID),
	)
	if err != nil {
		return err
	}
	//if there is no network error, bcs-api always response 200
	// we need to verify response body object
	// * if response is not json, low level http error
	// * if response object is Status, we can read Status.message
	// * Custom Resource Object when create successfully
	json, err := simplejson.NewJson(resp.Reply)
	if err != nil {
		return fmt.Errorf("response format is Not expected Json, http error: %s", string(resp.Reply))
	}
	replyKind, _ := json.Get("kind").String()
	if replyKind == StatusKind {
		msg, _ := json.Get("message").String()
		return fmt.Errorf("%s", msg)
	}
	return nil
}

//CreateResource create CRD by definition file
func (bs *bcsScheduler) CreateCustomResource(clusterID, apiVersion, plural, namespace string, data []byte) error {
	if apiVersion == "" {
		return fmt.Errorf("lost apiVersion for CustomResource")
	}
	if plural == "" {
		return fmt.Errorf("lost data type for CustomResource")
	}
	if namespace == "" {
		return fmt.Errorf("lost namespace for creation")
	}
	if len(data) == 0 {
		return fmt.Errorf("lost detail for creation")
	}
	baseURL := fmt.Sprintf(bcsSchedulerCustomResourceURL, bs.bcsAPIAddress)
	resp, err := bs.requester.DoForResponse(
		fmt.Sprintf("%s/%s/namespaces/%s/%s", baseURL, apiVersion, namespace, plural),
		http.MethodPost,
		data,
		getClusterIDHeader(clusterID),
	)
	if err != nil {
		return err
	}
	//if there is no network error, bcs-api always response 200
	// we need to verify response body object
	// * if response is not json, low level http error
	// * if response object is Status, we can read Status.message
	// * Custom Resource Object when create successfully
	json, err := simplejson.NewJson(resp.Reply)
	if err != nil {
		return fmt.Errorf("response format is Not expected Json, http error: %s", string(resp.Reply))
	}
	replyKind, _ := json.Get("kind").String()
	if replyKind == StatusKind {
		msg, _ := json.Get("message").String()
		return fmt.Errorf("%s", msg)
	}
	return nil
}

//UpdateResource replace specified CRD
func (bs *bcsScheduler) UpdateCustomResource(clusterID, apiVersion, plural, namespace, name string, data []byte) error {
	if apiVersion == "" {
		return fmt.Errorf("lost apiVersion for CustomResource")
	}
	if plural == "" {
		return fmt.Errorf("lost data type for CustomResource")
	}
	if len(data) == 0 {
		return fmt.Errorf("lost detail for creation")
	}
	oldData, err := bs.GetCustomResource(clusterID, apiVersion, plural, namespace, name)
	if err != nil {
		return err
	}
	//copy meta data for update
	oldObject, _ := simplejson.NewJson(oldData)
	updateObject, _ := simplejson.NewJson(data)
	oldMeta := oldObject.Get("metadata")
	updateObject.Set("metadata", oldMeta)
	updateData, _ := updateObject.Encode()
	//ready to Update
	baseURL := fmt.Sprintf(bcsSchedulerCustomResourceURL, bs.bcsAPIAddress)
	resp, err := bs.requester.DoForResponse(
		fmt.Sprintf("%s/%s/namespaces/%s/%s/%s", baseURL, apiVersion, namespace, plural, name),
		http.MethodPut,
		updateData,
		getClusterIDHeader(clusterID),
	)
	if err != nil {
		return err
	}
	//if there is no network error, bcs-api always response 200
	// we need to verify response body object
	// * if response is not json, low level http error
	// * if response object is Status, we can read Status.message
	// * Custom Resource Object when create successfully
	json, err := simplejson.NewJson(resp.Reply)
	if err != nil {
		return fmt.Errorf("response format is Not expected Json, http error: %s", string(resp.Reply))
	}
	replyKind, _ := json.Get("kind").String()
	if replyKind == StatusKind {
		msg, _ := json.Get("message").String()
		return fmt.Errorf("%s", msg)
	}
	return nil
}

//ListCustomResource list all created CRD
func (bs *bcsScheduler) ListCustomResource(clusterID, apiVersion, plural, namespace string) ([]byte, error) {
	baseURL := fmt.Sprintf(bcsSchedulerCustomResourceURL, bs.bcsAPIAddress)
	var reqURL string
	if namespace == AllNamespace {
		reqURL = fmt.Sprintf("%s/%s/%s", baseURL, apiVersion, plural)
	} else {
		reqURL = fmt.Sprintf("%s/%s/namespaces/%s/%s", baseURL, apiVersion, namespace, plural)
	}

	resp, err := bs.requester.DoForResponse(
		reqURL,
		http.MethodGet,
		nil,
		getClusterIDHeader(clusterID),
	)
	if err != nil {
		return nil, err
	}
	//if there is no network error, bcs-api always response 200
	// we need to verify response body object
	// * if response is not json, low level http error
	// * if response object is Status, we can read Status.message
	// * Custom Resource Object when create successfully
	json, err := simplejson.NewJson(resp.Reply)
	if err != nil {
		return nil, fmt.Errorf("response format is Not expected Json, http error: %s", string(resp.Reply))
	}
	replyKind, _ := json.Get("kind").String()
	if replyKind == StatusKind {
		msg, _ := json.Get("message").String()
		return nil, fmt.Errorf("%s", msg)
	}
	return resp.Reply, nil
}

//GetCustomResource get specified CRD
func (bs *bcsScheduler) GetCustomResource(clusterID, apiVersion, plural, namespace, name string) ([]byte, error) {
	baseURL := fmt.Sprintf(bcsSchedulerCustomResourceURL, bs.bcsAPIAddress)
	reqURL := fmt.Sprintf("%s/%s/namespaces/%s/%s/%s", baseURL, apiVersion, namespace, plural, name)
	resp, err := bs.requester.DoForResponse(
		reqURL,
		http.MethodGet,
		nil,
		getClusterIDHeader(clusterID),
	)
	if err != nil {
		return nil, err
	}
	//if there is no network error, bcs-api always response 200
	// we need to verify response body object
	// * if response is not json, low level http error
	// * if response object is Status, we can read Status.message
	// * Custom Resource Object when create successfully
	json, err := simplejson.NewJson(resp.Reply)
	if err != nil {
		return nil, fmt.Errorf("response format is Not expected Json, http error: %s", string(resp.Reply))
	}
	replyKind, _ := json.Get("kind").String()
	if replyKind == StatusKind {
		msg, _ := json.Get("message").String()
		return nil, fmt.Errorf("%s", msg)
	}
	return resp.Reply, nil
}

//DeleteCustomResource delete specified CRD
func (bs *bcsScheduler) DeleteCustomResource(clusterID, apiVersion, plural, namespace, name string) error {
	baseURL := fmt.Sprintf(bcsSchedulerCustomResourceURL, bs.bcsAPIAddress)
	reqURL := fmt.Sprintf("%s/%s/namespaces/%s/%s/%s", baseURL, apiVersion, namespace, plural, name)
	resp, err := bs.requester.DoForResponse(
		reqURL,
		http.MethodDelete,
		nil,
		getClusterIDHeader(clusterID),
	)
	if err != nil {
		return err
	}
	//if there is no network error, bcs-api always response 200
	// we need to verify response body object
	// * if response is not json, low level http error
	// * if response object is Status, we can read Status.message
	// * Custom Resource Object when create successfully
	json, err := simplejson.NewJson(resp.Reply)
	if err != nil {
		return fmt.Errorf("response format is Not expected Json, http error: %s", string(resp.Reply))
	}
	replyKind, _ := json.Get("kind").String()
	if replyKind == StatusKind {
		msg, _ := json.Get("message").String()
		if msg == "" {
			//success
			return nil
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}
