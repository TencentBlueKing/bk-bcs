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

package cache

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/cache"
	"github.com/stretchr/testify/mock"
	"time"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Set(key string, value interface{}, expiration time.Duration) (string, error) {
	args := m.Called(key, value, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockCache) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	args := m.Called(key, value, expiration)
	return args.Bool(0), args.Error(1)
}

func (m *MockCache) SetEX(key string, value interface{}, expiration time.Duration) (string, error) {
	args := m.Called(key, value, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockCache) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockCache) Del(key string) (uint64, error) {
	args := m.Called(key)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockCache) Expire(key string, expiration time.Duration) (bool, error) {
	args := m.Called(key, expiration)
	return args.Bool(0), args.Error(1)
}

var _ cache.Cache = new(MockCache)
