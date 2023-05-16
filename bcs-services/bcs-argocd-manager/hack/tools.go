//go:build tools
// +build tools

/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package hack

import (
	// k8s.io/code-generator is vendored to get generate-groups.sh, and k8s codegen utilities
	_ "github.com/gogo/protobuf/gogoproto"
	_ "github.com/golang/protobuf/protoc-gen-go"                       // protoc-gen-go TODO
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway" // protoc-gen-grpc-gateway TODO
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger"      // protoc-gen-swagger TODO
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"                  // protoc-gen-go-grpc TODO
	_ "k8s.io/api/core/v1"                                             // v1 TODO
	_ "k8s.io/client-go"                                               // client-go TODO
	_ "k8s.io/code-generator"                                          // code-generator TODO
	_ "k8s.io/code-generator/cmd/client-gen"                           // client-gen TODO
	_ "k8s.io/code-generator/cmd/deepcopy-gen"                         // deepcopy-gen TODO
	_ "k8s.io/code-generator/cmd/defaulter-gen"                        // defaulter-gen TODO
	_ "k8s.io/code-generator/cmd/go-to-protobuf"                       // go-to-protobuf TODO
	_ "k8s.io/code-generator/cmd/go-to-protobuf/protoc-gen-gogo"       // protoc-gen-gogo TODO
	_ "k8s.io/code-generator/cmd/informer-gen"                         // informer-gen TODO
	_ "k8s.io/code-generator/cmd/lister-gen"                           // lister-gen TODO
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"                // controller-gen TODO
)
