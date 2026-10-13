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
	"github.com/Tencent/bk-bcs/bcs-common/common/encryptv2" // nolint
	"github.com/jinzhu/gorm"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/metrics"
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
	GetProjectClients(projectCode string) []models.BcsClientUser
	GetAllClients() []models.BcsClientUser
	GetClient(projectCode, name string) *models.BcsClientUser
	CreateClientToken(token *models.BcsClientUser) error
	UpdateClientToken(projectCode, name string, updatedClient *models.BcsClient) error
	DeleteProjectClient(projectCode, name string) error
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
	start := time.Now()
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
			metrics.ReportMysqlSlowQueryMetrics("GetTokenByCondition", metrics.Query, metrics.ErrStatus, start)
			blog.Errorf("decrypt token failed, err %s", err.Error())
			return nil
		}
		metrics.ReportMysqlSlowQueryMetrics("GetTokenByCondition", metrics.Query, metrics.SucStatus, start)
		return &token
	}
	metrics.ReportMysqlSlowQueryMetrics("GetTokenByCondition", metrics.Query, metrics.SucStatus, start)
	return nil
}

// GetUserTokensByName get user tokens by username, return user tokens that is expired and not expired,
func (u *realTokenStore) GetUserTokensByName(name string) []models.BcsUser {
	var err error
	var tokens []models.BcsUser
	start := time.Now()
	u.db.Where(&models.BcsUser{Name: name}).Find(&tokens)
	for k, v := range tokens {
		tokens[k].UserToken, err = u.decryptToken(v.UserToken)
		if err != nil {
			blog.Errorf("decrypt token failed, err %s", err.Error())
			continue
		}
	}
	metrics.ReportMysqlSlowQueryMetrics("GetUserTokensByName", metrics.Query, metrics.SucStatus, start)
	return tokens
}

// CreateToken create new token
func (u *realTokenStore) CreateToken(token *models.BcsUser) error {
	var err error
	start := time.Now()
	token.UserToken, err = u.encryptToken(token.UserToken)
	if err != nil {
		metrics.ReportMysqlSlowQueryMetrics("CreateToken", metrics.Create, metrics.ErrStatus, start)
		return err
	}
	err = u.db.Create(token).Error
	metrics.ReportMysqlSlowQueryMetrics("CreateToken", metrics.Create, metrics.SucStatus, start)
	return err
}

// UpdateToken update token information
func (u *realTokenStore) UpdateToken(token, updatedToken *models.BcsUser) error {
	var err error
	start := time.Now()
	token.UserToken, err = u.encryptToken(token.UserToken)
	if err != nil {
		return err
	}
	updatedToken.UserToken, err = u.encryptToken(updatedToken.UserToken)
	if err != nil {
		return err
	}
	err = u.db.Model(token).Updates(*updatedToken).Error
	if err != nil {
		metrics.ReportMysqlSlowQueryMetrics("UpdateToken", metrics.Update, metrics.ErrStatus, start)
		return err
	}
	metrics.ReportMysqlSlowQueryMetrics("UpdateToken", metrics.Update, metrics.SucStatus, start)
	return nil
}

// DeleteToken delete user token
func (u *realTokenStore) DeleteToken(token string) error {
	var err error
	start := time.Now()
	token, err = u.encryptToken(token)
	if err != nil {
		return err
	}
	cond := &models.BcsUser{UserToken: token}
	err = u.db.Where(cond).Delete(&models.BcsUser{}).Error
	if err != nil {
		metrics.ReportMysqlSlowQueryMetrics("DeleteToken", metrics.Delete, metrics.ErrStatus, start)
		return err
	}
	metrics.ReportMysqlSlowQueryMetrics("DeleteToken", metrics.Delete, metrics.SucStatus, start)
	return nil
}

// CreateTemporaryToken create new temporary token
func (u *realTokenStore) CreateTemporaryToken(token *models.BcsTempToken) error {
	var err error
	start := time.Now()
	token.Token, err = u.encryptToken(token.Token)
	if err != nil {
		return err
	}
	err = u.db.Create(token).Error
	if err != nil {
		metrics.ReportMysqlSlowQueryMetrics("CreateTemporaryToken", metrics.Create, metrics.ErrStatus, start)
		return err
	}
	metrics.ReportMysqlSlowQueryMetrics("CreateTemporaryToken", metrics.Create, metrics.SucStatus, start)
	return nil
}

// GetTempTokenByCondition Query temp user by condition
func (u *realTokenStore) GetTempTokenByCondition(cond *models.BcsTempToken) *models.BcsTempToken {
	tempUser := models.BcsTempToken{}
	var err error
	start := time.Now()
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
	metrics.ReportMysqlSlowQueryMetrics("GetTempTokenByCondition", metrics.Query, metrics.SucStatus, start)
	return nil
}

// GetAllNotExpiredTokens get available user
func (u *realTokenStore) GetAllNotExpiredTokens() []models.BcsUser {
	var tokens []models.BcsUser
	start := time.Now()
	u.db.Where("expires_at > ?", time.Now()).Find(&tokens)
	for k, v := range tokens {
		token, err := u.decryptToken(v.UserToken)
		if err != nil {
			blog.Errorf("decrypt token failed, err %s", err.Error())
			continue
		}
		tokens[k].UserToken = token
	}
	metrics.ReportMysqlSlowQueryMetrics("GetAllNotExpiredTokens", metrics.Query, metrics.SucStatus, start)
	return tokens
}

// GetAllTokens get all tokens
func (u *realTokenStore) GetAllTokens() []models.BcsUser {
	var tokens []models.BcsUser
	start := time.Now()
	u.db.Find(&tokens)
	for k, v := range tokens {
		token, err := u.decryptToken(v.UserToken)
		if err != nil {
			blog.Errorf("decrypt token failed, err %s", err.Error())
			continue
		}
		tokens[k].UserToken = token
	}
	metrics.ReportMysqlSlowQueryMetrics("GetAllTokens", metrics.Query, metrics.SucStatus, start)
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

func (u *realTokenStore) GetAllClients() []models.BcsClientUser {
	var err error
	var tokens []models.BcsClientUser
	start := time.Now()
	// nolint goconst
	u.db.Raw(`SELECT c.project_code, c.name, c.manager, c.authority_user, c.created_by, c.created_at, c.updated_at, `+
		`u.expires_at, u.user_token, u.tenant_id FROM bcs_clients AS c JOIN bcs_users AS u on u.name = c.name WHERE u.user_type = ? `+
		`AND u.deleted_at IS NULL AND c.deleted_at IS NULL`, models.PlainUser).
		Find(&tokens)
	for k, v := range tokens {
		tokens[k].UserToken, err = u.decryptToken(v.UserToken)
		if err != nil {
			blog.Errorf("decrypt token failed, err %s", err.Error())
			continue
		}
	}
	metrics.ReportMysqlSlowQueryMetrics("GetAllClients", metrics.Query, metrics.SucStatus, start)
	return tokens
}

func (u *realTokenStore) GetProjectClients(projectCode string) []models.BcsClientUser {
	var err error
	var tokens []models.BcsClientUser
	start := time.Now()
	// nolint:lll
	u.db.Raw(`SELECT c.project_code, c.name, c.manager, c.authority_user, c.created_by, c.created_at, c.updated_at, `+
		`u.expires_at, u.user_token, u.tenant_id FROM bcs_clients AS c JOIN bcs_users AS u on u.name = c.name WHERE u.user_type = ? `+
		`AND u.deleted_at IS NULL AND c.deleted_at IS NULL AND c.project_code = ?`, models.PlainUser, projectCode).
		Find(&tokens)
	for k, v := range tokens {
		tokens[k].UserToken, err = u.decryptToken(v.UserToken)
		if err != nil {
			blog.Errorf("decrypt token failed, err %s", err.Error())
			continue
		}
	}
	metrics.ReportMysqlSlowQueryMetrics("GetProjectClients", metrics.Query, metrics.SucStatus, start)
	return tokens
}

func (u *realTokenStore) GetClient(projectCode, name string) *models.BcsClientUser {
	var client models.BcsClientUser
	start := time.Now()
	// nolint:lll
	u.db.Raw(`SELECT c.project_code, c.name, c.manager, c.authority_user, c.created_by, c.created_at, c.updated_at, `+
		`u.expires_at, u.user_token, u.tenant_id FROM bcs_clients AS c JOIN bcs_users AS u on u.name = c.name WHERE u.user_type = ? `+
		`AND u.deleted_at IS NULL AND c.deleted_at IS NULL AND c.project_code = ? AND c.name = ?`,
		models.PlainUser, projectCode, name).Scan(&client)
	metrics.ReportMysqlSlowQueryMetrics("GetClient", metrics.Query, metrics.SucStatus, start)
	return &client
}

func (u *realTokenStore) CreateClientToken(clientUser *models.BcsClientUser) error {
	var err error
	start := time.Now()
	clientUser.UserToken, err = u.encryptToken(clientUser.UserToken)
	if err != nil {
		return err
	}

	token := &models.BcsUser{
		Name:      clientUser.Name,
		UserType:  clientUser.UserType,
		UserToken: clientUser.UserToken,
		CreatedBy: clientUser.CreatedBy,
		ExpiresAt: clientUser.ExpiresAt,
		TenantID:  clientUser.TenantID,
	}
	client := &models.BcsClient{
		ProjectCode: clientUser.ProjectCode,
		Name:        clientUser.Name,
		Manager:     &clientUser.CreatedBy,
		CreatedBy:   clientUser.CreatedBy,
	}

	// 开启事务
	err = u.db.Transaction(func(tx *gorm.DB) error {
		// 创建 client token
		if err = tx.Create(token).Error; err != nil {
			return err
		}
		// 删除已有的 client
		cond := &models.BcsClient{ProjectCode: clientUser.ProjectCode, Name: clientUser.Name}
		if err = tx.Where(cond).Delete(&models.BcsClient{}).Error; err != nil {
			return err
		}
		// 创建平台账号
		if err = tx.Create(client).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		metrics.ReportMysqlSlowQueryMetrics("CreateClientToken", metrics.Create, metrics.ErrStatus, start)
		return err
	}
	metrics.ReportMysqlSlowQueryMetrics("CreateClientToken", metrics.Create, metrics.SucStatus, start)
	return nil
}

func (u *realTokenStore) UpdateClientToken(projectCode, name string, updatedClient *models.BcsClient) error {
	start := time.Now()
	err := u.db.Model(models.BcsClient{}).Where("project_code = ? and name = ?", projectCode, name).
		Updates(*updatedClient).Error
	if err != nil {
		metrics.ReportMysqlSlowQueryMetrics("UpdateClientToken", metrics.Update, metrics.ErrStatus, start)
		return err
	}
	metrics.ReportMysqlSlowQueryMetrics("UpdateClientToken", metrics.Update, metrics.SucStatus, start)
	return nil
}

func (u *realTokenStore) DeleteProjectClient(projectCode, name string) error {
	start := time.Now()
	err := u.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(&models.BcsUser{Name: name}).Delete(&models.BcsUser{}).Error; err != nil {
			return err
		}
		if err := tx.Where(&models.BcsClient{ProjectCode: projectCode, Name: name}).
			Delete(&models.BcsClient{}).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		metrics.ReportMysqlSlowQueryMetrics("DeleteProjectClient", metrics.Delete, metrics.ErrStatus, start)
		return err
	}
	metrics.ReportMysqlSlowQueryMetrics("DeleteProjectClient", metrics.Delete, metrics.SucStatus, start)
	return nil
}
