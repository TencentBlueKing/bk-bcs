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

package store

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	applicationpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
)

type appHistoryStore struct {
	num       int
	queues    []chan *v1alpha1.Application
	appClient applicationpkg.ApplicationServiceClient
	db        dao.Interface
}

func (s *appHistoryStore) init() {
	s.queues = make([]chan *v1alpha1.Application, 0, s.num)
	for i := 0; i < s.num; i++ {
		s.queues = append(s.queues, make(chan *v1alpha1.Application, 1000))
	}
	for i := 0; i < s.num; i++ {
		go func(i int) {
			s.handle(s.queues[i])
		}(i)
	}
}

func (s *appHistoryStore) handle(ch chan *v1alpha1.Application) {
	for item := range ch {
		if err := s.handleApplication(item); err != nil {
			blog.Errorf("[HistoryStore] handle application history store failed: %s", err.Error())
		}
	}
}

func (s *appHistoryStore) handleApplication(item *v1alpha1.Application) error {
	if len(item.Status.History) == 0 {
		blog.Warnf("[HistoryStore] application '%s' have not history", item.Name)
		return nil
	}
	history := item.Status.History.LastRevisionHistory()
	hm, err := s.db.GetApplicationHistoryManifest(item.Name, string(item.UID), history.ID)
	if err != nil {
		return errors.Wrapf(err, "get application history '%s/%s/%d' manifest failed",
			item.Name, item.UID, history.ID)
	}
	if hm != nil {
		return nil
	}
	applicationYaml, err := json.Marshal(item)
	if err != nil {
		return errors.Wrapf(err, "marshal application '%s' failed", item.Name)
	}
	source, err := json.Marshal(history.Source)
	if err != nil {
		return errors.Wrapf(err, "marshal application '%s' history source failed", item.Name)
	}
	sources, err := json.Marshal(history.Sources)
	if err != nil {
		return errors.Wrapf(err, "marshal application '%s' history sources failed", item.Name)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	resp, err := s.appClient.ManagedResources(ctx, &applicationpkg.ResourcesQuery{
		ApplicationName: &item.Name,
	})
	if err != nil {
		return errors.Wrapf(err, "get managed resources from application '%s' failed", item.Name)
	}
	managedResources, err := json.Marshal(resp.Items)
	if err != nil {
		return errors.Wrapf(err, "marshal managed resources from application '%s' failed", item.Name)
	}
	hm = &dao.ApplicationHistoryManifest{
		Project:                item.Spec.Project,
		Name:                   item.Name,
		ApplicationUID:         string(item.UID),
		ApplicationYaml:        string(applicationYaml),
		ManagedResources:       string(managedResources),
		Revision:               history.Revision,
		Revisions:              strings.Join(history.Revisions, ","),
		HistoryID:              history.ID,
		HistoryDeployStartedAt: history.DeployStartedAt.Time,
		HistoryDeployedAt:      history.DeployedAt.Time,
		HistorySource:          string(source),
		HistorySources:         string(sources),
	}
	if err = s.db.CreateApplicationHistoryManifest(hm); err != nil {
		return errors.Wrapf(err, "save application '%s/%s' history '%d' to db failed",
			item.Name, item.UID, history.ID)
	}
	blog.Infof("[HistoryStore] save application '%s/%s' history '%d' success", item.Name, item.UID, history.ID)
	return nil
}

func (s *appHistoryStore) enqueue(application v1alpha1.Application) {
	sum := 0
	for _, char := range application.Name {
		sum += int(char)
	}
	dot := sum % s.num
	s.queues[dot] <- &application
}
