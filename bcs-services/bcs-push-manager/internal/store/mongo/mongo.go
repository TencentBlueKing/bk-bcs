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

// Package mongo sub cluster store
package mongo

import (
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
)

const (
	tableNamePrefix        = "push_manager_"
	pushEventTableName     = "event"
	pushTemplateTableName  = "template"
	pushWhitelistTableName = "whitelist"
)

const (
	pushTemplateUniqueKey  = "template_id"
	pushWhitelistUniqueKey = "whitelist_id"
	pushEventUniqueKey     = "event_id"
	pushDomainKey          = "domain"
)

// Public public model set
type Public struct {
	TableName           string
	Indexes             []drivers.Index
	DB                  drivers.DB
	IsTableEnsured      bool
	IsTableEnsuredMutex sync.RWMutex
}

// Server server model set
type Server struct {
	*ModelPushEvent
	*ModelPushTemplate
	*ModelPushWhitelist
}

// NewServer create new server
func NewServer(db drivers.DB) *Server {
	return &Server{
		ModelPushEvent:     NewModelPushEvent(db),
		ModelPushTemplate:  NewModelPushTemplate(db),
		ModelPushWhitelist: NewModelPushWhitelist(db),
	}
}

// GetPushEventModel get push event model
func (s *Server) GetPushEventModel() *ModelPushEvent {
	return s.ModelPushEvent
}

// GetPushTemplateModel get push template model
func (s *Server) GetPushTemplateModel() *ModelPushTemplate {
	return s.ModelPushTemplate
}

// GetPushWhitelistModel get push whitelist model
func (s *Server) GetPushWhitelistModel() *ModelPushWhitelist {
	return s.ModelPushWhitelist
}
