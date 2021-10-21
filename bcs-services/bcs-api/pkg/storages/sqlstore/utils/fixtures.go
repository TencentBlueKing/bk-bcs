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

package utils

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/options"
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
)

// CreateBootstrapUsers create the bootstrap users, the bootstrap users can be defined in config files
func CreateBootstrapUsers(users []options.BootStrapUser) {
	for _, u := range users {
		user := m.User{
			Name:        u.Name,
			IsSuperUser: u.IsSuperUser,
		}
		// Query if user already exists
		userInDb := sqlstore.GetUserByCondition(&m.User{Name: user.Name})
		if userInDb != nil {
			blog.Infof("bootstrap user(%s) already exists, skip creating...", user.Name)
			continue
		}

		blog.Infof("Creating bootstrap user(%s)...", user.Name)
		err := sqlstore.CreateUser(&user)
		if err != nil {
			blog.Warnf("Unable to create bootstrap user %s: %s", user.Name, err.Error())
			continue
		}

		// Create initial session tokens
		for _, value := range u.Tokens {
			blog.Infof("Creating bootstrap user token(%s)...", value)
			byteToken, err := encrypt.DesDecryptFromBase([]byte(value))
			if err != nil {
				blog.Warnf("Unable to create bootstrap user token %s-%s: %s", user.Name, value, err.Error())
			}
			_, err = sqlstore.GetOrCreateUserToken(&user, m.UserTokenTypeSession, string(byteToken))
			if err != nil {
				blog.Warnf("Unable to create bootstrap user token %s-%s: %s", user.Name, value, err.Error())
				continue
			}
		}
	}
}
