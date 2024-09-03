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

// Package argo xxx
package argo

import (
	"context"
	"fmt"

	appv1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/util/db"
)

// Client argo client
type Client interface {
	GetRepository(ctx context.Context, repo string) (*appv1.Repository, error)
}

type client struct {
	argoDB db.ArgoDB
}

// New client
func New(argoDB db.ArgoDB) Client {
	return &client{argoDB: argoDB}
}

// GetRepository get repo
func (c *client) GetRepository(ctx context.Context, repo string) (*appv1.Repository, error) {
	repoInfo, err := c.argoDB.GetRepository(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("get repository %s failed:%s", repo, err.Error())
	}
	if repoInfo == nil {
		return nil, fmt.Errorf("get repository %s empty", repo)
	}
	return repoInfo, nil
}
