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

package analyze

import (
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
)

// CollectApplication defines the collect of application
type CollectApplication interface {
	ListApplicationCollects(project string) ([]*dao.ResourcePreference, error)
	ApplicationCollect(project, name string) error
	ApplicationCancelCollect(project, name string) error
}

type collectApplicationClient struct {
	db dao.Interface
}

// NewCollectApplication craete the client that application collect
func NewCollectApplication() CollectApplication {
	return &collectApplicationClient{
		db: dao.GlobalDB(),
	}
}

// ApplicationCollect will collect application
func (c *collectApplicationClient) ApplicationCollect(project, name string) error {
	if err := c.db.SaveResourcePreference(&dao.ResourcePreference{
		Project:      project,
		ResourceType: dao.PreferenceTypeApplication,
		Name:         name,
	}); err != nil {
		return errors.Wrapf(err, "application collect '%s/%s' failed", project, name)
	}
	return nil
}

// ApplicationCancelCollect will cancel collect application
func (c *collectApplicationClient) ApplicationCancelCollect(project, name string) error {
	if err := c.db.DeleteResourcePreference(project, dao.PreferenceTypeApplication, name); err != nil {
		return errors.Wrapf(err, "application cancel collect '%s/%s' failed", project, name)
	}
	return nil
}

// ListApplicationCollects return the application collects for project
func (c *collectApplicationClient) ListApplicationCollects(project string) ([]*dao.ResourcePreference, error) {
	prefers, err := c.db.ListResourcePreferences(project, dao.PreferenceTypeApplication)
	if err != nil {
		return nil, errors.Wrapf(err, "list preferences failed")
	}
	result := make([]*dao.ResourcePreference, 0, len(prefers))
	preferMap := make(map[string]*dao.ResourcePreference)
	for i := range prefers {
		preferMap[prefers[i].Name] = &prefers[i]
	}
	for _, prefer := range preferMap {
		result = append(result, prefer)
	}
	return result, nil
}
