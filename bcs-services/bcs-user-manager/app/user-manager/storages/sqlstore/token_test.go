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
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock
}

func (s *Suite) SetupSuite() {
	var (
		db  *sql.DB
		err error
	)
	db, s.mock, err = sqlmock.New()
	require.NoError(s.T(), err)

	s.DB, err = gorm.Open("postgres", db)
	require.NoError(s.T(), err)

	s.DB.LogMode(true)
}

func (s *Suite) AfterTest(_, _ string) {
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func TestToken(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) TestRealTokenStore_GetTokenByCondition() {
	tokenStore := NewTokenStore(s.DB)
	token := &models.BcsUser{
		Name: "test",
	}
	token1 := &models.BcsUser{
		ID:        1,
		Name:      "test",
		UserToken: "token",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	sqlGet := `SELECT * FROM "bcs_users"  WHERE "bcs_users"."deleted_at" IS NULL AND (("bcs_users"."name" = $1)) ORDER BY "bcs_users"."id" ASC LIMIT 1`
	s.mock.ExpectQuery(regexp.QuoteMeta(sqlGet)).
		WithArgs(token.Name).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "user_token", "created_at", "updated_at", "expires_at", "deleted_at"}).
		AddRow(token1.ID, token1.Name, token1.UserToken, token1.CreatedAt, token1.UpdatedAt, token1.ExpiresAt, nil))
	tokenInDB := tokenStore.GetTokenByCondition(token)
	require.NotNil(s.T(), tokenInDB)
	assert.Equal(s.T(), token1, tokenInDB)
}
