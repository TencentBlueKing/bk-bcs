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

package permission

import (
	"errors"
	"fmt"
	"net/http"
)

const (
	// Mesos cluster
	Mesos ClusterType = "mesos"
	// K8s cluster
	K8s ClusterType = "k8s"
)

// ClusterType cluster type
type ClusterType string

func (ct ClusterType) String() string {
	return string(ct)
}

// UserInfo userID/name
type UserInfo struct {
	UserID   uint
	UserName string
}

// ClusterResource cluster permission metadata
type ClusterResource struct {
	ClusterType ClusterType
	ClusterID   string
	Namespace   string
	URL         string
}

// VerifyClusterPermissionRequest cluster request permission
type VerifyClusterPermissionRequest struct {
	UserToken string `json:"user_token"`
	// ClusterType for k8s or mesos
	ClusterType ClusterType `json:"cluster_type"`
	ClusterID   string      `json:"cluster_id"`
	Namespace   string      `json:"namespace"`
	//ResourceNames string `json:"resource_names"`
	// URL for check not namespace resource
	RequestURL string `json:"request_url"`

	// Action for (POST GET PUT PATCH DELETE)
	Action string `json:"action"`
}

func (vpr *VerifyClusterPermissionRequest) validate() error {
	if vpr == nil {
		return errors.New("VerifyPermissionRequest is null")
	}

	if vpr.UserToken == "" {
		return errors.New("VerifyPermissionRequest user token is null")
	}

	switch vpr.Action {
	case http.MethodPost, http.MethodPut, http.MethodGet, http.MethodDelete, http.MethodPatch:
	default:
		return errors.New("VerifyPermissionRequest invalid action")
	}

	if vpr.ClusterType != K8s && vpr.ClusterType != Mesos {
		return fmt.Errorf("VerifyPermissionRequest invalid cluster_type[%s]", vpr.ClusterType)
	}

	// resourceType cluster or register system
	if vpr.ClusterID == "" {
		return fmt.Errorf("ClusterResource clusterID is null")
	}

	return nil
}

// VerifyServicePermissionRequest service request permission
type VerifyServicePermissionRequest struct {
	UserToken string `json:"user_token"`

	// ResourceType for service(clustermanager/usermanager)
	ResourceType string `json:"resource_type"`
	Resource     string `json:"resource"`

	// Action for (POST GET PUT PATCH DELETE)
	Action string `json:"action"`
}

func (vpr *VerifyServicePermissionRequest) validate() error {
	if vpr == nil {
		return errors.New("VerifyServicePermissionRequest is null")
	}

	if vpr.UserToken == "" {
		return errors.New("VerifyServicePermissionRequest user token is null")
	}

	switch vpr.Action {
	case http.MethodPost, http.MethodPut, http.MethodGet, http.MethodDelete, http.MethodPatch:
	default:
		return errors.New("VerifyServicePermissionRequest invalid action")
	}

	if vpr.ResourceType == "" || vpr.Resource == "" {
		return errors.New("VerifyServicePermissionRequest resource_type or resource is null")
	}

	// valid resource_type and resource for resource_type

	return nil
}

// VerifyPermissionReq for permission v2 request
type VerifyPermissionReq struct {
	UserToken    string `json:"user_token" validate:"required"`
	ResourceType string `json:"resource_type" validate:"required"`
	// clusterType mesos/k8s when ResourceType="cluster"
	ClusterType ClusterType `json:"cluster_type"`
	ClusterID   string      `json:"cluster_id"`
	RequestURL  string      `json:"request_url"`

	Resource string `json:"resource"`
	Action   string `json:"action" validate:"required"`
}

func (vpr *VerifyPermissionReq) validate() error {
	if vpr == nil {
		return errors.New("VerifyPermissionRequest is null")
	}

	if vpr.UserToken == "" {
		return errors.New("VerifyPermissionRequest user token is null")
	}

	switch vpr.Action {
	case http.MethodPost, http.MethodPut, http.MethodGet, http.MethodDelete, http.MethodPatch:
	default:
		return errors.New("VerifyPermissionRequest invalid action")
	}

	if vpr.ResourceType == "cluster" {
		if vpr.ClusterType != K8s && vpr.ClusterType != Mesos {
			return fmt.Errorf("VerifyPermissionRequest invalid cluster_type[%s]", vpr.ClusterType)
		}

		// resourceType cluster or register system
		if vpr.ClusterID == "" {
			return fmt.Errorf("ClusterResource clusterID is null")
		}
	}

	return nil
}
