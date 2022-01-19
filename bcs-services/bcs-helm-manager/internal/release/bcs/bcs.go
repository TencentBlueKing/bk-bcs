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

package bcs

import (
	"context"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release/bcs/sdk"
)

// New return a new release.Handler instance
func New(c release.Config) release.Handler {
	return &handler{
		config: &c,
		sdkClientGroup: sdk.NewGroup(sdk.Config{
			BcsAPI:         c.APIServer,
			Token:          c.Token,
			PatchTemplates: c.PatchTemplates,
			VarTemplates:   c.VarTemplates,
		}),
	}
}

type handler struct {
	config *release.Config

	sdkClientGroup sdk.Group
}

// Cluster get a cluster handler by clusterID
func (h *handler) Cluster(clusterID string) release.Cluster {
	return &cluster{
		handler:   h,
		clusterID: clusterID,
	}
}

type cluster struct {
	*handler

	clusterID string

	sdkClientLock sync.RWMutex
	sdkClientSet  sdk.Client
}

func (c *cluster) ensureSdkClient() sdk.Client {
	c.sdkClientLock.Lock()
	defer c.sdkClientLock.Unlock()

	if c.sdkClientSet != nil {
		return c.sdkClientSet
	}

	c.sdkClientSet = c.sdkClientGroup.Cluster(c.clusterID)
	return c.sdkClientSet
}

// List release
func (c *cluster) List(ctx context.Context, option release.ListOption) (int, []*release.Release, error) {
	return c.list(ctx, option)
}

// Install release
func (c *cluster) Install(ctx context.Context, conf release.HelmInstallConfig) (*release.HelmInstallResult, error) {
	return c.install(ctx, conf)
}

// Uninstall release
func (c *cluster) Uninstall(ctx context.Context, conf release.HelmUninstallConfig) (
	*release.HelmUninstallResult, error) {
	return c.uninstall(ctx, conf)
}

// Upgrade release
func (c *cluster) Upgrade(ctx context.Context, conf release.HelmUpgradeConfig) (*release.HelmUpgradeResult, error) {
	return c.upgrade(ctx, conf)
}

// Rollback release
func (c *cluster) Rollback(ctx context.Context, conf release.HelmRollbackConfig) (*release.HelmRollbackResult, error) {
	return c.rollback(ctx, conf)
}
