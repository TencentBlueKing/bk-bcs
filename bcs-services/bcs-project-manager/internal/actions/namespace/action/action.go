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

// Package action xxx
package action

import (
	"context"

	spb "google.golang.org/protobuf/types/known/structpb"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// NamespaceAction action for namespace
type NamespaceAction interface {
	SyncNamespace(ctx context.Context, req *proto.SyncNamespaceRequest, resp *proto.SyncNamespaceResponse) error
	WithdrawNamespace(ctx context.Context,
		req *proto.WithdrawNamespaceRequest, resp *proto.WithdrawNamespaceResponse) error
	CreateNamespace(ctx context.Context, req *proto.CreateNamespaceRequest, resp *proto.CreateNamespaceResponse) error
	CreateNamespaceCallback(ctx context.Context,
		req *proto.NamespaceCallbackRequest, resp *proto.NamespaceCallbackResponse) error
	UpdateNamespace(ctx context.Context,
		req *proto.UpdateNamespaceRequest, resp *proto.UpdateNamespaceResponse) error
	UpdateNamespaceCallback(ctx context.Context,
		req *proto.NamespaceCallbackRequest, resp *proto.NamespaceCallbackResponse) error
	GetNamespace(ctx context.Context, req *proto.GetNamespaceRequest, resp *proto.GetNamespaceResponse) error
	ListNamespaces(ctx context.Context, req *proto.ListNamespacesRequest, resp *proto.ListNamespacesResponse) error
	ListNativeNamespaces(ctx context.Context,
		req *proto.ListNativeNamespacesRequest, resp *proto.ListNativeNamespacesResponse) error
	ListNativeNamespacesContent(ctx context.Context,
		req *proto.ListNativeNamespacesContentRequest, resp *spb.Struct) error
	DeleteNamespace(ctx context.Context, req *proto.DeleteNamespaceRequest, resp *proto.DeleteNamespaceResponse) error
	DeleteNamespaceCallback(ctx context.Context,
		req *proto.NamespaceCallbackRequest, resp *proto.NamespaceCallbackResponse) error
}
