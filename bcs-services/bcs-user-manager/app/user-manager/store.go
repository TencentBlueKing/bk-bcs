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

// Package usermanager xxx
package usermanager

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"
	jwt2 "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/jwt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/cache"
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
		&models.BcsOperationLog{},
		&models.BcsTokenNotify{},
		&models.BcsTempToken{},
		&models.Activity{},
		&models.BcsClient{},
	)

	// remove user name Constraints, because we will soft delete token on db when user destroy there token,
	// so we can't use unique index to check user name
	sqlstore.GCoreDB.Model(&models.BcsUser{}).RemoveIndex("name")

	err := createBootstrapUsers(conf.BootStrapUsers)
	if err != nil {
		return err
	}

	go syncTokenToRedis()

	return nil
}

// createBootstrapUsers create the bootstrap users, the bootstrap users can be defined in config files
// NOCC:golint/fnsize(设计如此)
func createBootstrapUsers(users []options.BootStrapUser) error {
	tokenStore := sqlstore.NewTokenStore(sqlstore.GCoreDB, config.GlobalCryptor)
	for _, u := range users {
		var userType uint
		var subType jwt.UserType
		var expiresAt time.Time
		switch u.UserType {
		case "admin":
			userType = models.AdminUser
			subType = jwt.Client
			expiresAt = time.Now().Add(sqlstore.AdminSaasUserExpiredTime)
		case "saas":
			userType = models.SaasUser
			subType = jwt.Client
			expiresAt = time.Now().Add(sqlstore.AdminSaasUserExpiredTime)
		case "plain":
			userType = models.PlainUser
			subType = jwt.User
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
		} else {
			err = sqlstore.CreateUser(&user)
			if err != nil {
				return fmt.Errorf("error creating user [%s]: %s", user.Name, err.Error())
			}
		}

		// create user token
		tokenInDB := tokenStore.GetTokenByCondition(&models.BcsUser{Name: user.Name, UserType: user.UserType})
		if tokenInDB != nil {
			blog.Infof("bootstrap user(%s) token already exists, skip creating...", user.Name)
		} else {
			err = tokenStore.CreateToken(&models.BcsUser{
				Name:      user.Name,
				UserToken: user.UserToken,
				UserType:  user.UserType,
				CreatedBy: models.CreatedBySystem,
				ExpiresAt: expiresAt,
			})
			if err != nil {
				return fmt.Errorf("error creating user [%s] token: %s", user.Name, err.Error())
			}
		}

		// create user token jwt
		userInfo := &jwt.UserInfo{
			SubType:     subType.String(),
			ExpiredTime: int64(time.Until(user.ExpiresAt).Seconds()),
			Issuer:      jwt.JWTIssuer,
		}
		if userInfo.SubType == jwt.Client.String() {
			userInfo.ClientName = user.Name
		} else {
			userInfo.UserName = user.Name
		}
		jwtString, err := jwt2.JWTClient.JWTSign(userInfo)
		if err != nil {
			return fmt.Errorf("error creating jwt for user [%s]: %s", user.Name, err.Error())
		}
		// nolint
		_, err = cache.RDB.SetNX(context.TODO(), constant.TokenKeyPrefix+user.UserToken, jwtString, user.ExpiresAt.Sub(time.Now()))
		if err != nil {
			return fmt.Errorf("error storing user [%s] jwt: %s", user.Name, err.Error())
		}
	}
	return nil
}

// syncTokenToRedis will fetch user token from bcs_tokens, and store it to redis
func syncTokenToRedis() {
	tokenStore := sqlstore.NewTokenStore(sqlstore.GCoreDB, config.GlobalCryptor)
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	if !config.GetGlobalConfig().EnableTokenSync {
		syncToken(tokenStore)
		return
	}

	for {
		syncToken(tokenStore)
		// nolint
		select {
		case <-ticker.C:
		}
	}
}

func syncToken(tokenStore sqlstore.TokenStore) {
	tokens := tokenStore.GetAllNotExpiredTokens()
	blog.Infof("sync token to redis, total %d", len(tokens))
	done := 0
	needLess := 0
	for _, v := range tokens {
		// create user token jwt
		userInfo := &jwt.UserInfo{
			ExpiredTime: int64(time.Until(v.ExpiresAt).Seconds()),
			Issuer:      jwt.JWTIssuer,
		}
		if v.IsClient() {
			userInfo.SubType = jwt.Client.String()
			userInfo.ClientName = v.Name
		} else {
			userInfo.SubType = jwt.User.String()
			userInfo.UserName = v.Name
		}
		jwtString, err := jwt2.JWTClient.JWTSign(userInfo)
		if err != nil {
			blog.Errorf("error creating jwt for user [%s]: %s", v.Name, err.Error())
			continue
		}
		set, err := cache.RDB.SetNX(context.TODO(), constant.TokenKeyPrefix+v.UserToken, jwtString, time.Until(v.ExpiresAt)) // nolint
		if err != nil {
			blog.Errorf("error storing user [%s] jwt: %s", v.Name, err.Error())
		}
		if set {
			done++
		} else {
			needLess++
		}
	}

	// sync client authorize token
	clients := tokenStore.GetAllClients()
	for _, v := range clients {
		for _, user := range v.AuthorityUserList() {
			userInfo := &jwt.UserInfo{
				SubType:     jwt.User.String(),
				UserName:    user,
				ExpiredTime: int64(time.Until(v.ExpiresAt).Seconds()),
				Issuer:      jwt.JWTIssuer,
			}
			jwtString, err := jwt2.JWTClient.JWTSign(userInfo)
			if err != nil {
				blog.Errorf("error creating jwt for user [%s]: %s", v.Name, err.Error())
				continue
			}
			key := fmt.Sprintf("%sop-%s:%s", constant.TokenKeyPrefix, user, v.UserToken)
			_, err = cache.RDB.SetNX(context.TODO(), key, jwtString, time.Until(v.ExpiresAt))
			if err != nil {
				blog.Errorf("error storing user [%s] jwt: %s", v.Name, err.Error())
			}
		}
	}
	blog.Infof("sync %d token to redis done, %d token not need to sync", done, needLess)
}
