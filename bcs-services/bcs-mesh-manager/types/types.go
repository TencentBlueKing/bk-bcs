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

package types

const (
	// IstioOperatorKind CRD kind for istio operator
	IstioOperatorKind string = "IstioOperator"
	// IstioOperatorGroup CRD group info
	IstioOperatorGroup string = "install.istio.io"
	// IstioOperatorVersion version for CRD
	IstioOperatorVersion string = "v1alpha1"
	// IstioOperatorName stable name for istio
	IstioOperatorName string = "istiocontrolplane"
	// IstioOperatorNamespace namespace for istio install
	IstioOperatorNamespace string = "istio-system"
	// IstioOperatorPlural plural info for meshManager request with API
	IstioOperatorPlural string = "istiooperators"
	// IstioOperatorListKind list kind for operator
	IstioOperatorListKind string = "IstioOperatorList"
)
