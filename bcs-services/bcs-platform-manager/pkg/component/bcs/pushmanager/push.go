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

// Package pushmanager xxx
package pushmanager

import (
	"context"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/pushmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
)

// PushEvents push events
func PushEvents(ctx context.Context, event *pushmanager.PushEvent) (string, error) {
	cli, close, err := pushmanager.GetClient(config.ServiceDomain)
	if err != nil {
		return "", err
	}

	defer Close(close)

	resp, err := cli.CreatePushEvent(ctx, &pushmanager.CreatePushEventRequest{
		Domain: event.Domain,
		Event:  event,
	})
	if err != nil {
		return "", err
	}

	return resp.EventId, nil
}

// GetPushTemplate get push template
func GetPushTemplate(ctx context.Context, domain, templateID string) (*pushmanager.TemplateContent, error) {
	cli, close, err := pushmanager.GetClient(config.ServiceDomain)
	if err != nil {
		return nil, err
	}

	defer Close(close)

	resp, err := cli.GetPushTemplate(ctx, &pushmanager.GetPushTemplateRequest{
		Domain:     domain,
		TemplateId: templateID,
	})
	if err != nil {
		return nil, err
	}

	return resp.Template.Content, nil
}

// ListPushTemplateByID list push template by id
func ListPushTemplateByID(ctx context.Context, domain, templateID string) ([]*pushmanager.PushTemplate, error) {
	cli, close, err := pushmanager.GetClient(config.ServiceDomain)
	if err != nil {
		return nil, err
	}

	defer Close(close)

	resp, err := cli.ListPushTemplates(ctx, &pushmanager.ListPushTemplatesRequest{
		Domain: domain,
	})
	if err != nil {
		return nil, err
	}

	templates := make([]*pushmanager.PushTemplate, 0)
	for _, template := range resp.Templates {
		if strings.Contains(template.TemplateId, templateID) {
			templates = append(templates, template)
		}
	}

	return templates, nil
}
