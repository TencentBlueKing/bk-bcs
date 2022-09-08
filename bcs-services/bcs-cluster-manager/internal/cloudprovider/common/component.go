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

package common

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/component/watch"
)

// EnsureWatchComponentTask deploy bcs-k8s-watch task, if not exist, create it, if exist, update it
func EnsureWatchComponentTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	// get auto scaling option
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	projectID := step.Params[cloudprovider.ProjectIDKey.String()]

	// InstallWatchComponentByHelm install watch component but not handle error, need user to handle release
	err = InstallWatchComponentByHelm(ctx, projectID, clusterID)
	if err != nil {
		blog.Errorf("EnsureWatchComponentTask[%s] failed: %v", taskID, err)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("EnsureAutoScalerTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func InstallWatchComponentByHelm(ctx context.Context, projectID, clusterID string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	bcsWatch := &watch.BcsWatch{
		ClusterID: clusterID,
	}
	values, err := bcsWatch.GetValues()
	if err != nil {
		blog.Errorf("InstallWatchComponentByHelm[%s] get bcsWatch[%s] failed: %v", taskID, clusterID, err)
		return err
	}
	blog.Infof("InstallWatchComponentByHelm[%s] get bcsWatchValues[%s] successful", taskID, values)

	installer, err := watch.GetWatchInstaller(projectID)
	if err != nil {
		blog.Errorf("InstallWatchComponentByHelm[%s] GetWatchInstaller failed: %v", taskID, err)
		return err
	}
	err = installer.Install(clusterID, values)
	if err != nil {
		blog.Errorf("InstallWatchComponentByHelm[%s] Install failed: %v", err)
		return err
	}

	blog.Infof("InstallWatchComponentByHelm[%s] successful[%s:%s]", taskID, projectID, clusterID)
	return nil
}

// DeleteWatchComponentByHelm unInstall watch
func DeleteWatchComponentByHelm(ctx context.Context, projectID, clusterID string) error {
	traceID := cloudprovider.GetTaskIDFromContext(ctx)

	install, err := watch.GetWatchInstaller(projectID)
	if err != nil {
		blog.Errorf("DeleteWatchComponentByHelm[%s] GetWatchInstaller failed: %v", traceID, err)
		return err
	}
	err = install.Uninstall(clusterID)
	if err != nil {
		blog.Errorf("DeleteWatchComponentByHelm[%s] Uninstall failed: %v", err)
		return err
	}

	blog.Infof("DeleteWatchComponentByHelm[%s] successful[%s:%s]", traceID, projectID, clusterID)
	return nil
}
