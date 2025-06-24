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
 */

package k8sclient

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// BkLogConfig represents a BkLogConfig custom resource
type BkLogConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              *BkLogConfigSpec `json:"spec,omitempty"`
}

// BkLogConfigSpec defines the desired state of BkLogConfig
type BkLogConfigSpec struct {
	DataID             int                   `json:"dataId,omitempty"`
	LogConfigType      string                `json:"logConfigType"`
	Namespace          string                `json:"namespace"`
	ContainerNameMatch []string              `json:"containerNameMatch,omitempty"`
	LabelSelector      *metav1.LabelSelector `json:"labelSelector,omitempty"`
	Path               []string              `json:"path,omitempty"`
	Encoding           string                `json:"encoding,omitempty"`
	ExtMeta            map[string]string     `json:"extMeta,omitempty"`
}

// BKLogConfigGVR is the GroupVersionResource for BkLogConfig
var BKLogConfigGVR = schema.GroupVersionResource{
	Group:    "bk.tencent.com",
	Version:  "v1alpha1",
	Resource: "bklogconfigs",
}

// GetBkLogConfig retrieves a specific BkLogConfig custom resource from the specified cluster, namespace, and name.
func GetBkLogConfig(ctx context.Context, clusterID string, namespace string, name string) (*BkLogConfig, error) {
	dynamicClient, err := GetDynamicClientByClusterId(clusterID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get dynamic client for cluster %s", clusterID)
	}

	unstructuredObj, err := dynamicClient.Resource(BKLogConfigGVR).Namespace(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get BkLogConfig %s/%s in cluster %s",
			namespace, name, clusterID)
	}

	rawJSON, err := unstructuredObj.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal unstructured BkLogConfig to JSON")
	}

	bkLogConfig := &BkLogConfig{}
	if err := json.Unmarshal(rawJSON, bkLogConfig); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal BkLogConfig from JSON")
	}

	return bkLogConfig, nil
}

// CreateBkLogConfig creates a new BkLogConfig custom resource.
func CreateBkLogConfig(
	ctx context.Context, clusterID string, namespace string, config *BkLogConfig,
) (*BkLogConfig, error) {
	dynamicClient, err := GetDynamicClientByClusterId(clusterID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get dynamic client for cluster %s", clusterID)
	}

	config.TypeMeta = metav1.TypeMeta{
		Kind:       "BkLogConfig",
		APIVersion: "bk.tencent.com/v1alpha1",
	}
	config.ObjectMeta.Namespace = namespace

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal BkLogConfig to JSON")
	}

	unstructuredObj := &unstructured.Unstructured{}
	if uerr := json.Unmarshal(configJSON, unstructuredObj); uerr != nil {
		return nil, errors.Wrap(uerr, "failed to unmarshal BkLogConfig JSON to unstructured")
	}

	createdObj, err := dynamicClient.Resource(BKLogConfigGVR).Namespace(namespace).
		Create(ctx, unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create BkLogConfig %s/%s in cluster %s",
			namespace, config.Name, clusterID)
	}

	createdJSON, err := createdObj.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal created unstructured BkLogConfig to JSON")
	}

	createdBkLogConfig := &BkLogConfig{}
	if err := json.Unmarshal(createdJSON, createdBkLogConfig); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal created BkLogConfig from JSON")
	}

	return createdBkLogConfig, nil
}

// UpdateBkLogConfig updates an existing BkLogConfig custom resource.
// The provided config object should have its ResourceVersion field set for optimistic concurrency control.
func UpdateBkLogConfig(ctx context.Context, clusterID string, namespace string, config *BkLogConfig,
) (*BkLogConfig, error) {
	dynamicClient, err := GetDynamicClientByClusterId(clusterID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get dynamic client for cluster %s", clusterID)
	}

	if config.ObjectMeta.ResourceVersion == "" {
		return nil, errors.New("failed to update BkLogConfig: ResourceVersion must be set")
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal BkLogConfig to JSON for update")
	}

	unstructuredObj := &unstructured.Unstructured{}
	err = json.Unmarshal(configJSON, unstructuredObj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal BkLogConfig JSON to unstructured for update")
	}

	updatedObj, err := dynamicClient.Resource(BKLogConfigGVR).Namespace(namespace).
		Update(ctx, unstructuredObj, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update BkLogConfig %s/%s in cluster %s",
			namespace, config.Name, clusterID)
	}

	updatedJSON, err := updatedObj.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal updated unstructured BkLogConfig to JSON")
	}

	updatedBkLogConfig := &BkLogConfig{}
	err = json.Unmarshal(updatedJSON, updatedBkLogConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal updated BkLogConfig from JSON")
	}

	return updatedBkLogConfig, nil
}

// DeleteBkLogConfig deletes a BkLogConfig custom resource.
func DeleteBkLogConfig(ctx context.Context, clusterID string, namespace string, name string) error {
	dynamicClient, err := GetDynamicClientByClusterId(clusterID)
	if err != nil {
		return errors.Wrapf(err, "failed to get dynamic client for cluster %s", clusterID)
	}

	err = dynamicClient.Resource(BKLogConfigGVR).Namespace(namespace).
		Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to delete BkLogConfig %s/%s in cluster %s",
			namespace, name, clusterID)
	}

	return nil
}
