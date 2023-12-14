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

// Package auth NOTES
package auth

import (
	"context"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"
)

// Authorizer defines all the supported functionalities to do auth operation.
type Authorizer interface {
	// Authorize check if a user's operate resource is already authorized or not.
	Authorize(ctx context.Context, opts *client.AuthOptions) (*client.Decision, error)

	// AuthorizeBatch check if a user's operate resources is authorized or not batch.
	// Note: being authorized resources must be the same resource.
	AuthorizeBatch(ctx context.Context, opts *client.AuthBatchOptions) ([]*client.Decision, error)

	// AuthorizeAnyBatch check if a user have any authority of the operate actions batch.
	AuthorizeAnyBatch(ctx context.Context, opts *client.AuthBatchOptions) ([]*client.Decision, error)

	// ListAuthorizedInstances list a user's all the authorized resource instance list with an action.
	// Note: opts.Resources are not required.
	// the returned list may be huge, we do not do result paging
	ListAuthorizedInstances(ctx context.Context, opts *client.AuthOptions, resourceType client.TypeID) (
		*client.AuthorizeList, error)

	// GrantResourceCreatorAction grant a user's resource creator action.
	GrantResourceCreatorAction(ctx context.Context, opts *client.GrantResourceCreatorActionOption) error
}

// ResourceFetcher defines all the supported operations for iam to fetch resources from bscp
type ResourceFetcher interface {
	// ListInstancesWithAttributes get "same" resource instances with attributes
	// returned with the resource's instance id list matched with options.
	ListInstancesWithAttributes(ctx context.Context, opts *client.ListWithAttributes) (idList []string, err error)
}

// NewAuth initialize an authorizer
func NewAuth(c *client.Client, fetcher ResourceFetcher) (Authorizer, error) {
	if c == nil {
		return nil, errf.New(errf.InvalidParameter, "client is nil")
	}

	if fetcher == nil {
		return nil, errf.New(errf.InvalidParameter, "fetcher is nil")
	}

	return &Authorize{
		client:  c,
		fetcher: fetcher,
	}, nil
}
