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

// Api versions allow the api contract for a resource to be changed while keeping
// backward compatibility by support multiple concurrent versions
// of the same resource

//go:generate deepcopy-gen -O zz_generated.deepcopy -i . -h ../../../../boilerplate.go.txt
//go:generate defaulter-gen -O zz_generated.defaults -i . -h ../../../../boilerplate.go.txt
//go:generate conversion-gen -O zz_generated.conversion -i . -h ../../../../boilerplate.go.txt

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen=github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis/aggregation
// +k8s:defaulter-gen=TypeMeta
// +groupName=aggregation.federated.bkbcs.tencent.com
package v1alpha1 // import "github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis/aggregation/v1alpha1"
