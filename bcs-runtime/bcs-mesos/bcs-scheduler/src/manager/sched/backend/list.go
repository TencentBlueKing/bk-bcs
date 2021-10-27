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

package backend

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

func (b *backend) ListApplications(runAs string) ([]*types.Application, error) {
	return b.store.ListApplications(runAs)
}

func (b *backend) ListApplicationTaskGroups(runAs, appId string) ([]*types.TaskGroup, error) {
	b.store.LockApplication(runAs + "." + appId)
	defer b.store.UnLockApplication(runAs + "." + appId)

	return b.store.ListTaskGroups(runAs, appId)
}

// ListApplicationVersions is used to list all versions for application from db specified by application id.
func (b *backend) ListApplicationVersions(runAs, appId string) ([]string, error) {
	return b.store.ListVersions(runAs, appId)
}
