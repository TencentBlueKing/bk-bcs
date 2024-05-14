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

package kubernetes

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LockRecord lock record
type LockRecord struct {
	// OwnerID client id for this lock
	OwnerID string `json:"ownerID"`
	// ExpireDuration lock expired after this duration
	ExpireDuration time.Duration `json:"expireDuration"`
	// AcquireTime acquire time
	AcquireTime metav1.Time `json:"acquireTime"`
	// RenewTime renew time
	RenewTime metav1.Time `json:"renewTime"`
	// ResourceVersion resource version for k8s object
	ResourceVersion string `json:"-"`
}
