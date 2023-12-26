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

// Package auth NOTES
package auth

import (
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/lcache"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// New initialize the auth service instance.
func New(cache *lcache.Cache) *AuthService {
	return &AuthService{
		cache: cache,
	}
}

// AuthService defines auth related operations.
type AuthService struct { //nolint:revive
	cache *lcache.Cache
}

// Authorize if user has permission to the bscp resource.
func (as *AuthService) Authorize(kt *kit.Kit, res *meta.ResourceAttribute) (bool, error) {
	return as.cache.Auth.Authorize(kt, res)
}

// CanMatchCI if credential can match the config item.
func (as *AuthService) CanMatchCI(kt *kit.Kit, bizID uint32,
	app string, token string, path string, name string) (bool, error) {
	return as.cache.Credential.CanMatchCI(kt, bizID, app, token, path, name)
}
