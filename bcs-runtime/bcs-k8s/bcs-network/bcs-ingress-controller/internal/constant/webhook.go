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

package constant

const (
	// ValidateMsgUnknownErr ingress validate get unknown error
	ValidateMsgUnknownErr = "unknown err: %+v"
	// ValidateMsgEmptySvc ingress validate get empty service
	ValidateMsgEmptySvc = "rule[%d] should have service"
	// ValidateMsgNotFoundSvc ingress validate can not get specified service
	ValidateMsgNotFoundSvc = "rule[%d] service '%s/%s' not found"
	// ValidateRouteMsgEmptySvc ingress validate get empty service
	ValidateRouteMsgEmptySvc = "rule[%d] route[%d] should have service"
	// ValidateRouteMsgNotFoundSvc ingress validate can not get specified service
	ValidateRouteMsgNotFoundSvc = "rule[%d] route[%d] service '%s/%s' not found"
	// ValidateMsgInvalidWorkload ingress validate get invalid workload
	ValidateMsgInvalidWorkload = "port mapping[%d]'s workload have empty workload kind/namespace/name "
	// ValidateMsgEmptyWorkload ingress validate get empty workload
	ValidateMsgEmptyWorkload = "port mapping[%d]'s workload not found"

	// PortConflictMsg new create ingress/portpool has port conflict with existed
	PortConflictMsg = "port conflict with kind[%s] namespace[%s] name[%s] on lbID[%s]"
)
