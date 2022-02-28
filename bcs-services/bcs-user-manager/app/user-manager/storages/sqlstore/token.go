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
	"time"

	"github.com/jinzhu/gorm"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
)

type TokenStore interface {
	GetTokenByCondition(cond *models.BcsToken) *models.BcsToken
	GetUserTokensByName(name string) []models.BcsToken
	CreateToken(token *models.BcsToken) error
	UpdateToken(token, updatedToken *models.BcsToken) error
	DeleteToken(token string) error
	CreateTemporaryToken(token *models.BcsTempToken) error
}

func NewTokenStore(db *gorm.DB) TokenStore {
	return &realTokenStore{db: db}
}

type realTokenStore struct {
	db *gorm.DB
}

// GetTokenByCondition Query token by condition
func (u *realTokenStore) GetTokenByCondition(cond *models.BcsToken) *models.BcsToken {
	token := models.BcsToken{}
	u.db.Where(cond).First(&token)
	if token.ID != 0 {
		return &token
	}
	return nil
}

// GetUserTokensByName get user tokens by username, return user tokens that is expired and not expired,
func (u *realTokenStore) GetUserTokensByName(name string) []models.BcsToken {
	var tokens []models.BcsToken
	u.db.Where(&models.BcsToken{Username: name}).Find(&tokens)
	return tokens
}

// CreateToken create new token
func (u *realTokenStore) CreateToken(token *models.BcsToken) error {
	err := u.db.Create(token).Error
	return err
}

// UpdateToken update token information
func (u *realTokenStore) UpdateToken(token, updatedToken *models.BcsToken) error {
	err := u.db.Model(token).Updates(*updatedToken).Error
	return err
}

// DeleteToken delete user token
func (u *realTokenStore) DeleteToken(token string) error {
	cond := &models.BcsToken{Token: token}
	err := u.db.Where(cond).Delete(&models.BcsToken{}).Error
	return err
}

// CreateToken create new temporary token
func (u *realTokenStore) CreateTemporaryToken(token *models.BcsTempToken) error {
	err := u.db.Create(token).Error
	return err
}

func GetAllNotExpiredTokens() []models.BcsToken {
	var tokens []models.BcsToken
	GCoreDB.Where("expires_at > ?", time.Now()).Find(&tokens)
	return tokens
}

func GetAllTokens() []models.BcsToken {
	var tokens []models.BcsToken
	GCoreDB.Find(&tokens)
	return tokens
}
