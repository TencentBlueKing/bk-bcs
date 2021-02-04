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

package usermanager

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/options"
)

// SetupStore setup db
func SetupStore(conf *config.UserMgrConfig) error {
	if err := sqlstore.InitCoreDatabase(conf); err != nil {
		return fmt.Errorf("error initing database: %s", err.Error())
	}

	// Migrate db schemas
	sqlstore.GCoreDB.AutoMigrate(
		&models.BcsUser{},
		&models.BcsCluster{},
		&models.BcsRegisterToken{},
		&models.BcsClusterCredential{},
		&models.BcsRole{},
		&models.BcsUserResourceRole{},
		&models.TkeCidr{},
		&models.BcsWsClusterCredentials{},
	)

	err := createBootstrapUsers(conf.BootStrapUsers)
	if err != nil {
		return err
	}

	return nil
}

// createBootstrapUsers create the bootstrap users, the bootstrap users can be defined in config files
func createBootstrapUsers(users []options.BootStrapUser) error {
	for _, u := range users {
		var userType uint
		var expiresAt time.Time
		switch u.UserType {
		case "admin":
			userType = sqlstore.AdminUser
			expiresAt = time.Now().Add(sqlstore.AdminSaasUserExpiredTime)
		case "saas":
			userType = sqlstore.SaasUser
			expiresAt = time.Now().Add(sqlstore.AdminSaasUserExpiredTime)
		case "plain":
			userType = sqlstore.PlainUser
			expiresAt = time.Now().Add(sqlstore.PlainUserExpiredTime)
		default:
			return fmt.Errorf("invalid user type, user type must be [admin, saas, plain]")
		}
		byteToken, err := encrypt.DesDecryptFromBase([]byte(u.Token))
		if err != nil {
			return fmt.Errorf("error decrypting token for user [%s], %s", u.Name, err.Error())
		}
		user := models.BcsUser{
			Name:      u.Name,
			UserType:  userType,
			UserToken: string(byteToken),
			ExpiresAt: expiresAt,
		}

		// Query if user already exists
		userInDb := sqlstore.GetUserByCondition(&models.BcsUser{Name: user.Name, UserType: user.UserType})
		if userInDb != nil {
			blog.Infof("bootstrap user(%s) already exists, skip creating...", user.Name)
			continue
		}

		err = sqlstore.CreateUser(&user)
		if err != nil {
			return fmt.Errorf("error creating user [%s]: %s", user.Name, err.Error())
		}
	}
	return nil
}
