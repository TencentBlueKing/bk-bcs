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

package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/operation"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
)

// ReleaseUninstallAction release uninstall action
type ReleaseUninstallAction struct {
	model          store.HelmManagerModel
	platform       repo.Platform // nolint
	releaseHandler release.Handler

	clusterID string
	name      string
	namespace string
	username  string
}

// ReleaseUninstallActionOption options
type ReleaseUninstallActionOption struct {
	Model          store.HelmManagerModel
	ReleaseHandler release.Handler

	ClusterID string
	Name      string
	Namespace string
	Username  string
}

// NewReleaseUninstallAction new release uninstall action
func NewReleaseUninstallAction(o *ReleaseUninstallActionOption) *ReleaseUninstallAction {
	return &ReleaseUninstallAction{
		model:          o.Model,
		releaseHandler: o.ReleaseHandler,
		clusterID:      o.ClusterID,
		name:           o.Name,
		namespace:      o.Namespace,
		username:       o.Username,
	}
}

var _ operation.Operation = &ReleaseUninstallAction{}

// Action xxx
func (r *ReleaseUninstallAction) Action() string {
	return "Uninstall"
}

// Name xxx
func (r *ReleaseUninstallAction) Name() string {
	return fmt.Sprintf("uninstall-%s", r.name)
}

// Prepare xxx
func (r *ReleaseUninstallAction) Prepare(ctx context.Context) error {
	return nil
}

// Validate xxx
func (r *ReleaseUninstallAction) Validate() error {
	return nil
}

// Execute xxx
func (r *ReleaseUninstallAction) Execute(ctx context.Context) error {
	// check release revision exist
	_, err := r.releaseHandler.Cluster(r.clusterID).Get(ctx, release.GetOption{
		Namespace: r.namespace, Name: r.name,
	})
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return nil
		}
		blog.Warnf("get %s/%s in cluster %s error, %s", r.namespace, r.name, r.clusterID, err.Error())
		return nil
	}

	_, err = r.releaseHandler.Cluster(r.clusterID).Uninstall(ctx, release.HelmUninstallConfig{
		Namespace: r.namespace,
		Name:      r.name,
	})
	if err != nil {
		return fmt.Errorf("uninstall %s/%s in cluster %s error, %s",
			r.namespace, r.name, r.clusterID, err.Error())
	}
	return nil
}

// Done xxx
func (r *ReleaseUninstallAction) Done(err error) {
	if err != nil {
		rl := entity.M{
			entity.FieldKeyUpdateBy: r.username,
			entity.FieldKeyStatus:   common.ReleaseStatusUninstallFailed,
			entity.FieldKeyMessage:  err.Error(),
		}
		_ = r.model.UpdateRelease(context.Background(), r.clusterID, r.namespace, r.name, rl)
		return
	}
	_ = r.model.DeleteRelease(context.Background(), r.clusterID, r.namespace, r.name)
}
