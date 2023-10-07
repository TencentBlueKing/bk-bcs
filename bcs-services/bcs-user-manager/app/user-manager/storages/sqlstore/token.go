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

package sqlstore

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encryptv2"
	"github.com/jinzhu/gorm"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
)

// TokenStore is the token store that operate token in database
type TokenStore interface {
	GetTokenByCondition(cond *models.BcsUser) *models.BcsUser
	GetUserTokensByName(name string) []models.BcsUser
	CreateToken(token *models.BcsUser) error
	UpdateToken(token, updatedToken *models.BcsUser) error
	DeleteToken(token string) error
	CreateTemporaryToken(token *models.BcsTempToken) error
	GetTempTokenByCondition(cond *models.BcsTempToken) *models.BcsTempToken
	GetAllNotExpiredTokens() []models.BcsUser
	GetAllTokens() []models.BcsUser
}

// NewTokenStore create new token store with db
func NewTokenStore(db *gorm.DB, cryptor encryptv2.Cryptor) TokenStore {
	return &realTokenStore{db: db, cryptor: cryptor}
}

type realTokenStore struct {
	db      *gorm.DB
	cryptor encryptv2.Cryptor
}

// GetTokenByCondition Query token by condition
func (u *realTokenStore) GetTokenByCondition(cond *models.BcsUser) *models.BcsUser {
	var err error
	token := models.BcsUser{}
	if cond.UserToken != "" {
		cond.UserToken, err = u.encryptToken(cond.UserToken)
		if err != nil {
			blog.Errorf("encrypt token failed, err %s", err.Error())
			return nil
		}
	}

	u.db.Where(cond).First(&token)
	if token.ID != 0 {
		token.UserToken, err = u.decryptToken(token.UserToken)
		if err != nil {
			blog.Errorf("decrypt token failed, err %s", err.Error())
			return nil
		}
		return &token
	}
	return nil
}

// GetUserTokensByName get user tokens by username, return user tokens that is expired and not expired,
func (u *realTokenStore) GetUserTokensByName(name string) []models.BcsUser {
	var err error
	var tokens []models.BcsUser
	u.db.Where(&models.BcsUser{Name: name}).Find(&tokens)
	for k, v := range tokens {
		tokens[k].UserToken, err = u.decryptToken(v.UserToken)
		if err != nil {
			blog.Errorf("decrypt token failed, err %s", err.Error())
			continue
		}
	}
	return tokens
}

// CreateToken create new token
func (u *realTokenStore) CreateToken(token *models.BcsUser) error {
	var err error
	token.UserToken, err = u.encryptToken(token.UserToken)
	if err != nil {
		return err
	}
	err = u.db.Create(token).Error
	return err
}

// UpdateToken update token information
func (u *realTokenStore) UpdateToken(token, updatedToken *models.BcsUser) error {
	var err error
	token.UserToken, err = u.encryptToken(token.UserToken)
	if err != nil {
		return err
	}
	updatedToken.UserToken, err = u.encryptToken(updatedToken.UserToken)
	if err != nil {
		return err
	}
	err = u.db.Model(token).Updates(*updatedToken).Error
	return err
}

// DeleteToken delete user token
func (u *realTokenStore) DeleteToken(token string) error {
	var err error
	token, err = u.encryptToken(token)
	if err != nil {
		return err
	}
	cond := &models.BcsUser{UserToken: token}
	err = u.db.Where(cond).Delete(&models.BcsUser{}).Error
	return err
}

// CreateTemporaryToken create new temporary token
func (u *realTokenStore) CreateTemporaryToken(token *models.BcsTempToken) error {
	var err error
	token.Token, err = u.encryptToken(token.Token)
	if err != nil {
		return err
	}
	err = u.db.Create(token).Error
	return err
}

// GetTempTokenByCondition Query temp user by condition
func (u *realTokenStore) GetTempTokenByCondition(cond *models.BcsTempToken) *models.BcsTempToken {
	tempUser := models.BcsTempToken{}
	var err error
	if cond.Token != "" {
		cond.Token, err = u.encryptToken(cond.Token)
		if err != nil {
			blog.Errorf("encrypt token failed, err %s", err.Error())
			return nil
		}
	}
	u.db.Where(cond).First(&tempUser)
	if tempUser.ID != 0 {
		tempUser.Token, err = u.decryptToken(tempUser.Token)
		if err != nil {
			blog.Errorf("decrypt token failed, err %s", err.Error())
			return nil
		}
		return &tempUser
	}
	return nil
}

// GetAllNotExpiredTokens get available user
func (u *realTokenStore) GetAllNotExpiredTokens() []models.BcsUser {
	var tokens []models.BcsUser
	u.db.Where("expires_at > ?", time.Now()).Find(&tokens)
	for k, v := range tokens {
		token, err := u.decryptToken(v.UserToken)
		if err != nil {
			blog.Errorf("decrypt token failed, err %s", err.Error())
			continue
		}
		tokens[k].UserToken = token
	}
	return tokens
}

// GetAllTokens get all tokens
func (u *realTokenStore) GetAllTokens() []models.BcsUser {
	var tokens []models.BcsUser
	u.db.Find(&tokens)
	for k, v := range tokens {
		token, err := u.decryptToken(v.UserToken)
		if err != nil {
			blog.Errorf("decrypt token failed, err %s", err.Error())
			continue
		}
		tokens[k].UserToken = token
	}
	return tokens
}

func (u *realTokenStore) encryptToken(token string) (string, error) {
	if u.cryptor == nil {
		return token, nil
	}
	return u.cryptor.Encrypt(token)
}

func (u *realTokenStore) decryptToken(token string) (string, error) {
	// if token is not encrypted, return directly
	if len(token) == constant.DefaultTokenLength {
		return token, nil
	}
	if u.cryptor == nil {
		return token, nil
	}
	return u.cryptor.Decrypt(token)
}
