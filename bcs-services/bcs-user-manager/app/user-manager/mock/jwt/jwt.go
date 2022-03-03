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

package jwt

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/stretchr/testify/mock"
)

type MockJWTClient struct {
	mock.Mock
}

func (m *MockJWTClient) JWTSign(user *jwt.UserInfo) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *MockJWTClient) JWTDecode(jwtToken string) (*jwt.UserClaimsInfo, error) {
	args := m.Called(jwtToken)
	return args.Get(0).(*jwt.UserClaimsInfo), args.Error(1)
}

var _ jwt.BCSJWTAuthentication = &MockJWTClient{}
