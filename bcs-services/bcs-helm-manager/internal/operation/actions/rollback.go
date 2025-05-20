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
	"fmt"

	helmrelease "helm.sh/helm/v3/pkg/release"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/operation"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
)

// ReleaseRollbackAction release rollback action
type ReleaseRollbackAction struct {
	model          store.HelmManagerModel
	releaseHandler release.Handler

	clusterID string
	name      string
	namespace string
	revision  int
	username  string
}

// ReleaseRollbackActionOption options
type ReleaseRollbackActionOption struct {
	Model          store.HelmManagerModel
	ReleaseHandler release.Handler

	ClusterID string
	Name      string
	Namespace string
	Revision  int
	Username  string
}

// NewReleaseRollbackAction new release rollback action
func NewReleaseRollbackAction(o *ReleaseRollbackActionOption) *ReleaseRollbackAction {
	return &ReleaseRollbackAction{
		model:          o.Model,
		releaseHandler: o.ReleaseHandler,
		clusterID:      o.ClusterID,
		name:           o.Name,
		namespace:      o.Namespace,
		revision:       o.Revision,
		username:       o.Username,
	}
}

var _ operation.Operation = &ReleaseRollbackAction{}

// Action xxx
func (r *ReleaseRollbackAction) Action() string {
	return "Rollback"
}

// Name xxx
func (r *ReleaseRollbackAction) Name() string {
	return fmt.Sprintf("rollback-%s", r.name)
}

// Prepare xxx
func (r *ReleaseRollbackAction) Prepare(ctx context.Context) error {
	return nil
}

// Validate xxx
func (r *ReleaseRollbackAction) Validate(ctx context.Context) error {
	return nil
}

// Execute xxx
func (r *ReleaseRollbackAction) Execute(ctx context.Context) error {
	_, err := r.releaseHandler.Cluster(r.clusterID).Rollback(
		ctx,
		release.HelmRollbackConfig{
			Name:      r.name,
			Namespace: r.namespace,
			Revision:  r.revision,
		})
	if err != nil {
		return fmt.Errorf("rollback %s/%s in cluster %s to revision %d error, %s",
			r.namespace, r.name, r.clusterID, r.revision, err.Error())
	}
	return nil
}

// Done xxx
func (r *ReleaseRollbackAction) Done(err error) {
	status := helmrelease.StatusDeployed
	message := ""
	if err != nil {
		status = common.ReleaseStatusRollbackFailed
		message = err.Error()
	}
	rl := entity.M{
		entity.FieldKeyUpdateBy: r.username,
		entity.FieldKeyStatus:   status.String(),
		entity.FieldKeyMessage:  message,
	}
	detail, _ := r.releaseHandler.Cluster(r.clusterID).Get(context.Background(), release.GetOption{
		Namespace: r.namespace,
		Name:      r.name,
	})
	if detail != nil {
		rl.Update(entity.FieldKeyChartVersion, detail.ChartVersion)
		rl.Update(entity.FieldKeyRevision, detail.Revision)
		rl.Update(entity.FieldKeyValues, []string{detail.Values})
	}
	_ = r.model.UpdateRelease(context.Background(), r.clusterID, r.namespace, r.name, rl)
}
