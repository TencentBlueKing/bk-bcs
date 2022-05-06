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

package store

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/stretchr/testify/mock"
)

// MockTokenNotifyStore is a mock of TokenNotifyStore
type MockTokenNotifyStore struct {
	mock.Mock
}

// CreateTokenNotify mock create token notify
func (m *MockTokenNotifyStore) CreateTokenNotify(notify *models.BcsTokenNotify) error {
	args := m.Called(notify)
	return args.Error(0)
}

// GetTokenNotifyByCondition mock get token notify by condition
func (m *MockTokenNotifyStore) GetTokenNotifyByCondition(cond *models.BcsTokenNotify) []models.BcsTokenNotify {
	args := m.Called(cond)
	return args.Get(0).([]models.BcsTokenNotify)
}

// DeleteTokenNotify mock delete token notify
func (m *MockTokenNotifyStore) DeleteTokenNotify(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

var _ sqlstore.TokenNotifyStore = &MockTokenNotifyStore{}
