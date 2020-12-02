/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pluginutil

import (
	"encoding/json"
	"fmt"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sunstruct "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// AssertPod to check if the raw object is Pod type
func AssertPod(rawObject []byte) (bool, error) {
	tmpObject := &k8sunstruct.Unstructured{}
	if err := json.Unmarshal(rawObject, &tmpObject); err != nil {
		blog.Errorf("decode %s to unstructured object failed, err %s", string(rawObject), err.Error())
		return false, fmt.Errorf("decode data to unstructured object failed, err %s", err.Error())
	}
	if tmpObject.GroupVersionKind().Kind != "Pod" {
		blog.Warnf("object %s/%s is not Pod", tmpObject.GetNamespace(), tmpObject.GetName())
		return false, nil
	}
	return true, nil
}

// ToAdmissionResponse convert error to admission response
func ToAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
}
