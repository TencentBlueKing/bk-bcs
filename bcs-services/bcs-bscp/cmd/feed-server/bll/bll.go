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

// Package bll business logical layer which implements all the
// logical operations.
package bll

import (
	"fmt"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/auth"
	clientset "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/client-set"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/eventc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/lcache"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/observer"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/release"
	iamauth "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/auth"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
)

// New create a new BLL instance.
func New(sd serviced.Discover, authorizer iamauth.Authorizer, name string) (*BLL, error) {
	client, err := clientset.New(sd, authorizer)
	if err != nil {
		return nil, fmt.Errorf("new client set failed, err: %v", err)
	}

	localCache, err := lcache.NewLocalCache(client)
	if err != nil {
		return nil, err
	}

	obHandler := &observer.Handler{
		LocalCache: localCache.Purge,
	}
	ob, err := observer.New(obHandler, client, name)
	if err != nil {
		return nil, fmt.Errorf("new observer failed, err: %v", err)
	}

	schOpt := &eventc.Option{
		Observer: ob,
		Cache:    localCache,
	}
	sch, err := eventc.NewScheduler(schOpt, name)
	if err != nil {
		return nil, fmt.Errorf("new scheduler failed, err: %v", err)
	}

	rs, err := release.New(client, localCache, sch)
	if err != nil {
		return nil, fmt.Errorf("new release service failed, err: %v", err)
	}

	handler := &eventc.Handler{
		GetMatchedRelease: rs.GetMatchedRelease,
	}
	if err := sch.Run(handler); err != nil {
		return nil, fmt.Errorf("run scheduler faield, err: %v", err)
	}

	return &BLL{
		client:  client,
		release: rs,
		auth:    auth.New(localCache),
		cache:   localCache,
		ob:      ob,
		sch:     sch,
	}, nil
}

// BLL defines business logical layer instance.
type BLL struct {
	client  *clientset.ClientSet
	release *release.ReleasedService
	auth    *auth.AuthService
	cache   *lcache.Cache
	ob      observer.Interface
	sch     *eventc.Scheduler
}

// Release return the release service instance.
func (b *BLL) Release() *release.ReleasedService {
	return b.release
}

// Auth return the auth service instance.
func (b *BLL) Auth() *auth.AuthService {
	return b.auth
}

// AppCache return the app cache instance.
func (b *BLL) AppCache() *lcache.App {
	return b.cache.App
}

// RKvCache return the ReleasedKv cache instance.
func (b *BLL) RKvCache() *lcache.ReleasedKv {
	return b.cache.ReleasedKv
}

// ClientMetric return the client metric instance.
func (b *BLL) ClientMetric() *lcache.ClientMetric {
	return b.cache.ClientMetric
}
