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

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

//CreateResourceDefinition create CRD by definition file
func (bs *bcsScheduler) CreateCustomResourceDefinition(clusterID string, data []byte) error {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsScheudlerCustomResourceDefinitionURL, bs.bcsAPIAddress),
		http.MethodPost,
		data,
		getClusterIDHeader(clusterID),
	)
	if err != nil {
		return err
	}

	return nil
}

//UpdateResourceDefinition replace specified CRD
func (bs *bcsScheduler) UpdateCustomResourceDefinition(clusterID, name string, data []byte) error {

}

//ListCustomResourceDefinition list all created CRD
func (bs *bcsScheduler) ListCustomResourceDefinition(clusterID string) (*v1beta1.CustomResourceDefinitionList, error) {

}

//GetCustomResourceDefinition get specified CRD
func (bs *bcsScheduler) GetCustomResourceDefinition(clusterID string, name string) (*v1beta1.CustomResourceDefinition, error) {

}

//DeleteCustomResourceDefinition delete specified CRD
func (bs *bcsScheduler) DeleteCustomResourceDefinition(clusterID, name string) error {

}

//CreateResource create CRD by definition file
func (bs *bcsScheduler) CreateCustomResource(clusterID, namespace string, data []byte) error {

}

//UpdateResource replace specified CRD
func (bs *bcsScheduler) UpdateCustomResource(clusterID, namespace, name string, data []byte) error {

}

//ListCustomResource list all created CRD
func (bs *bcsScheduler) ListCustomResource(clusterID, namespace string) ([]byte, error) {

}

//GetCustomResource get specified CRD
func (bs *bcsScheduler) GetCustomResource(clusterID, namespace, name string) ([]byte, error) {

}

//DeleteCustomResource delete specified CRD
func (bs *bcsScheduler) DeleteCustomResource(clusterID, namespace, name string) error {

}
