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

type MockTokenStore struct {
	mock.Mock
}

func (m *MockTokenStore) GetTokenByCondition(cond *models.BcsUser) *models.BcsUser {
	args := m.Called(cond)
	return args.Get(0).(*models.BcsUser)
}

func (m *MockTokenStore) GetUserTokensByName(name string) []models.BcsUser {
	args := m.Called(name)
	return args.Get(0).([]models.BcsUser)
}

func (m *MockTokenStore) CreateToken(token *models.BcsUser) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockTokenStore) UpdateToken(token, updatedToken *models.BcsUser) error {
	args := m.Called(token, updatedToken)
	return args.Error(0)
}

func (m *MockTokenStore) DeleteToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockTokenStore) CreateTemporaryToken(token *models.BcsTempToken) error {
	args := m.Called(token)
	return args.Error(0)
}

var _ sqlstore.TokenStore = &MockTokenStore{}
