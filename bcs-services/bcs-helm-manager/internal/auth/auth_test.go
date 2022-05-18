/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package auth

import (
	"fmt"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/stretchr/testify/assert"
)

var adminName = "admin"

func newClient() *jwt.JWTClient {
	jwtOpt := jwt.JWTOptions{
		VerifyKeyFile: "./test/app.rsa.pub",
		SignKeyFile:   "./test/app.rsa",
	}
	jwtClient, err := jwt.NewJWTClient(jwtOpt)
	if err != nil {
		panic(fmt.Errorf("new jwt client error, %v", err))
	}
	return jwtClient
}

func newJWTToken() string {
	jwtClient := newClient()
	jwtToken, err := jwtClient.JWTSign(&jwt.UserInfo{
		ClientName:  "test",
		SubType:     jwt.User.String(),
		UserName:    adminName,
		ExpiredTime: 10000,
	})
	if err != nil {
		panic(fmt.Errorf("new jwt token error, %v", err))
	}
	return jwtToken
}

func TestParseUsername(t *testing.T) {
	jwtToken := newJWTToken()
	jwtClient, _ = jwt.NewJWTClient(
		jwt.JWTOptions{
			VerifyKeyFile: "./test/app.rsa.pub",
			SignKeyFile:   "./test/app.rsa",
		},
	)
	userAuth, _ := ParseUserFromJWT(jwtToken)
	assert.Equal(t, userAuth.Username, adminName)
}
