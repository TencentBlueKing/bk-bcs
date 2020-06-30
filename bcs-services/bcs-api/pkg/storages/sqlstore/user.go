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

package sqlstore

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
	"github.com/dchest/uniuri"
	"time"
)

const (
	// expired after 24 hours
	UserTokenForKubeconfigExpiredTime = 24 * time.Hour
	// this means never expired
	UserTokenForSessionExpiredTime = 10 * 365 * 24 * time.Hour
)

// Query user by user_id
func GetUser(id uint) *m.User {
	user := m.User{}
	GCoreDB.Where(&m.User{ID: id}).First(&user)
	if user.ID != 0 {
		return &user
	}
	return nil
}

// Query user by condition
func GetUserByCondition(cond *m.User) *m.User {
	user := m.User{}
	GCoreDB.Where(cond).First(&user)
	if user.ID != 0 {
		return &user
	}
	return nil
}

func CreateUser(user *m.User) error {
	err := GCoreDB.Create(user).Error
	return err
}

func CreateUserToken(userToken *m.UserToken) error {
	err := GCoreDB.Create(userToken).Error
	return err
}

func UpdateUserToken(userToken, updatedUserToken *m.UserToken) error {
	err := GCoreDB.Model(userToken).Updates(*updatedUserToken).Error
	return err
}

func GetUserToken(token string) *m.UserToken {
	if token == "" {
		return nil
	}
	userToken := m.UserToken{}
	GCoreDB.Where(&m.UserToken{Value: token}).First(&userToken)
	if userToken.ID != 0 {
		return &userToken
	}
	return nil
}

func GetExternalUserRecord(sourceType uint, userId, userType string) *m.ExternalUserRecord {
	if userId == "" || userType == "" {
		return nil
	}

	externalUserRecord := m.ExternalUserRecord{}
	GCoreDB.Where(&m.ExternalUserRecord{SourceType: sourceType, SourceUserType: userType, SourceUserId: userId}).First(&externalUserRecord)
	if externalUserRecord.ID != 0 {
		return &externalUserRecord
	}
	return nil

}

func CreateExternalUserRecord(record *m.ExternalUserRecord) error {
	err := GCoreDB.Create(record).Error
	return err
}

func GetOrCreateUser(sourceType uint, userId, userType string) (*m.User, error) {
	var user *m.User
	externalUserRecord := GetExternalUserRecord(sourceType, userId, userType)
	if externalUserRecord != nil {
		user = GetUser(externalUserRecord.UserId)
	} else {
		// create a new user
		user = &m.User{
			Name:        fmt.Sprintf("%s:%s", userType, userId),
			IsSuperUser: false,
		}
		err := CreateUser(user)
		if err != nil {
			return user, fmt.Errorf("CREATE_USER_FAIL: %s", err.Error())
		}
		eur := &m.ExternalUserRecord{
			UserId:         user.ID,
			SourceType:     sourceType,
			SourceUserType: userType,
			SourceUserId:   userId,
		}
		err = CreateExternalUserRecord(eur)
		if err != nil {
			return user, fmt.Errorf("CREATE_USER_FAIL: %s", err.Error())
		}
	}

	return user, nil
}

// GetUserTokenByUser get user_token of tokenType by user
func GetUserTokenByUser(user *m.User, tokenType uint) *m.UserToken {
	if user == nil {
		return nil
	}
	userToken := m.UserToken{}
	GCoreDB.Where(&m.UserToken{UserId: user.ID, Type: tokenType}).First(&userToken)
	if userToken.ID != 0 {
		return &userToken
	}
	return nil
}

const DefaultTokenLength = 32

func GetOrCreateUserToken(user *m.User, tokenType uint, defaultToken string) (*m.UserToken, error) {
	var userToken *m.UserToken
	userToken = GetUserTokenByUser(user, tokenType)
	if defaultToken == "" {
		defaultToken = uniuri.NewLen(DefaultTokenLength)
	}
	if userToken == nil {
		// create a new user_token
		if tokenType == m.UserTokenTypeKubeConfigForPaas {
			userToken = &m.UserToken{
				UserId:    user.ID,
				Type:      tokenType,
				Value:     defaultToken,
				ExpiresAt: time.Now().Add(UserTokenForKubeconfigExpiredTime),
			}
		} else if tokenType == m.UserTokenTypeSession || tokenType == m.UserTokenTypeKubeConfigPlain {
			userToken = &m.UserToken{
				UserId:    user.ID,
				Type:      tokenType,
				Value:     defaultToken,
				ExpiresAt: time.Now().Add(UserTokenForSessionExpiredTime),
			}
		}

		err := CreateUserToken(userToken)
		if err != nil {
			blog.Warnf("Unable to create user token %s: %s", user.Name, err.Error())
			return userToken, fmt.Errorf("CREATE_USER_TOKEN_FAIL: %s", err.Error())
		}
	} else if userToken != nil && tokenType == m.UserTokenTypeKubeConfigForPaas {
		updatedUserToken := userToken
		if time.Now().After(userToken.ExpiresAt) {
			updatedUserToken.Value = defaultToken
			updatedUserToken.ExpiresAt = time.Now().Add(UserTokenForKubeconfigExpiredTime)
		} else {
			updatedUserToken.ExpiresAt = time.Now().Add(UserTokenForKubeconfigExpiredTime)
		}
		err := UpdateUserToken(userToken, updatedUserToken)
		if err != nil {
			blog.Warnf("Unable to update user token %s: %s", user.Name, err.Error())
			return userToken, fmt.Errorf("UPDATE_USER_TOKEN_FAIL: %s", err.Error())
		}
		userToken = updatedUserToken
	}

	return userToken, nil
}
