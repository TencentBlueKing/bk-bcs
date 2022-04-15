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

// TokenNotifyStore is an interface of token notify store
type TokenNotifyStore interface {
	CreateTokenNotify(notify *models.BcsTokenNotify) error
	GetTokenNotifyByCondition(cond *models.BcsTokenNotify) []models.BcsTokenNotify
	DeleteTokenNotify(token string) error
}

// NewTokenNotifyStore creates a new token notify store
func NewTokenNotifyStore(db *gorm.DB) TokenNotifyStore {
	return &realTokenNotifyStore{db: db}
}

type realTokenNotifyStore struct {
	db *gorm.DB
}

// CreateTokenNotify create a new token notify
func (t *realTokenNotifyStore) CreateTokenNotify(notify *models.BcsTokenNotify) error {
	err := t.db.Create(notify).Error
	return err
}

// ExpireToken specify the response of token
type ExpireToken struct {
	Username  string    `json:"username"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GetTokenNotifyByCondition get token that has expired and not notified
func (t *realTokenNotifyStore) GetTokenNotifyByCondition(cond *models.BcsTokenNotify) []models.BcsTokenNotify {
	token := make([]models.BcsTokenNotify, 0)
	t.db.Where(cond).Find(&token)
	return token
}

// DeleteTokenNotify delete token notify
func (t *realTokenNotifyStore) DeleteTokenNotify(token string) error {
	err := t.db.Where("token = ?", token).Delete(&models.BcsTokenNotify{}).Error
	return err
}
