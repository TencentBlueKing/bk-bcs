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

package token

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/mock/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/mock/jwt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/mock/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/emicklei/go-restful"
)

type testToken struct {
	request              *restful.Request
	response             *restful.Response
	recorder             *httptest.ResponseRecorder
	tokenHandler         *TokenHandler
	mockTokenStore       *store.MockTokenStore
	mockTokenNotifyStore *store.MockTokenNotifyStore
	mockCache            *cache.MockCache
	mockJWTClient        *jwt.MockJWTClient
}

func newTestToken(form CreateTokenForm) (*testToken, error) {
	formData, err := json.Marshal(form)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(formData)
	r := httptest.NewRequest("POST", "/any", body)
	r.Header.Add("Content-Type", "application/json")
	request := restful.NewRequest(r)
	request.SetAttribute(constant.CurrentUserAttr, &models.BcsUser{Name: form.Username, UserType: sqlstore.PlainUser})
	recorder := httptest.NewRecorder()
	response := restful.NewResponse(recorder)
	response.SetRequestAccepts("application/json")
	mockTokenStore := new(store.MockTokenStore)
	mockTokenNotifyStore := new(store.MockTokenNotifyStore)
	mockCache := new(cache.MockCache)
	mockJWTClient := new(jwt.MockJWTClient)
	h := NewTokenHandler(mockTokenStore, mockTokenNotifyStore, mockCache, mockJWTClient)
	return &testToken{
		request:              request,
		response:             response,
		recorder:             recorder,
		tokenHandler:         h,
		mockTokenStore:       mockTokenStore,
		mockTokenNotifyStore: mockTokenNotifyStore,
		mockCache:            mockCache,
		mockJWTClient:        mockJWTClient,
	}, nil
}

func TestCreateToken(t *testing.T) {
	t.Run("invalid form", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 0}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		tt.tokenHandler.CreateToken(tt.request, tt.response)
		res := &utils.ErrorResponse{}
		err = json.Unmarshal(tt.recorder.Body.Bytes(), res)
		require.NoError(t, err)

		assert.Equal(t, 400, tt.response.StatusCode())
		assert.Equal(t, common.BcsErrApiBadRequest, res.Code)
	})

	t.Run("admin user create token", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 3600}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		tt.request.SetAttribute(constant.CurrentUserAttr, &models.BcsUser{Name: "admin", UserType: sqlstore.AdminUser})
		tt.mockTokenStore.On("GetUserTokensByName", form.Username).Once().
			Return([]models.BcsUser{}, nil)
		tt.mockCache.On("Set", mock.Anything, mock.Anything, time.Duration(form.Expiration)*time.Second).Once().
			Return("", nil)
		tt.mockJWTClient.On("JWTSign", mock.Anything).Once().Return("", nil)
		tt.mockTokenStore.On("CreateToken", mock.Anything).Once().Return(nil)
		tt.tokenHandler.CreateToken(tt.request, tt.response)
		res := &utils.ErrorResponse{}
		err = json.Unmarshal(tt.recorder.Body.Bytes(), res)
		require.NoError(t, err)

		tt.mockTokenStore.AssertExpectations(t)
		tt.mockTokenNotifyStore.AssertExpectations(t)
		tt.mockCache.AssertExpectations(t)
		tt.mockJWTClient.AssertExpectations(t)
		assert.Equal(t, 200, tt.response.StatusCode())
		assert.Equal(t, 0, res.Code)
		assert.Equal(t, "success", res.Message)
	})

	t.Run("not allow to access token", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 3600}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		tt.request.SetAttribute(constant.CurrentUserAttr, &models.BcsUser{Name: "user1", UserType: sqlstore.PlainUser})
		tt.tokenHandler.CreateToken(tt.request, tt.response)
		res := &utils.ErrorResponse{}
		err = json.Unmarshal(tt.recorder.Body.Bytes(), res)
		require.NoError(t, err)

		assert.Equal(t, 401, tt.response.StatusCode())
		assert.Equal(t, common.BcsErrApiUnauthorized, res.Code)
	})

	t.Run("user has token and token is not expired", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 3600}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		tt.mockTokenStore.On("GetUserTokensByName", form.Username).Once().
			Return([]models.BcsUser{{Name: "test", UserToken: "token", ExpiresAt: time.Now().Add(time.Hour)}}, nil)
		tt.tokenHandler.CreateToken(tt.request, tt.response)
		res := &utils.ErrorResponse{}
		err = json.Unmarshal(tt.recorder.Body.Bytes(), res)
		require.NoError(t, err)

		tt.mockTokenStore.AssertExpectations(t)
		tt.mockTokenNotifyStore.AssertExpectations(t)
		tt.mockCache.AssertExpectations(t)
		assert.Equal(t, 500, tt.response.StatusCode())
		assert.Equal(t, common.BcsErrApiInternalDbError, res.Code)
		message := fmt.Sprintf("errcode: %d, token already exists", common.BcsErrApiInternalDbError)
		assert.Equal(t, message, res.Message)
	})

	t.Run("user token is expired", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 3600}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		tt.mockTokenStore.On("GetUserTokensByName", form.Username).Once().
			Return([]models.BcsUser{{Name: "test", UserToken: "token", ExpiresAt: time.Now().Add(-1)}}, nil)
		tt.tokenHandler.CreateToken(tt.request, tt.response)
		res := &utils.ErrorResponse{}
		err = json.Unmarshal(tt.recorder.Body.Bytes(), res)
		require.NoError(t, err)

		tt.mockTokenStore.AssertExpectations(t)
		assert.Equal(t, 500, tt.response.StatusCode())
		assert.Equal(t, common.BcsErrApiInternalDbError, res.Code)
		message := fmt.Sprintf("errcode: %d, token already exists", common.BcsErrApiInternalDbError)
		assert.Equal(t, message, res.Message)
	})

	t.Run("user hasn't token", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 3600}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		tt.mockTokenStore.On("GetUserTokensByName", form.Username).Once().
			Return([]models.BcsUser{}, nil)
		tt.mockCache.On("Set", mock.Anything, mock.Anything, time.Duration(form.Expiration)*time.Second).Once().
			Return("", nil)
		tt.mockJWTClient.On("JWTSign", mock.Anything).Once().Return("", nil)
		tt.mockTokenStore.On("CreateToken", mock.Anything).Once().Return(nil)
		tt.tokenHandler.CreateToken(tt.request, tt.response)
		res := &utils.ErrorResponse{}
		err = json.Unmarshal(tt.recorder.Body.Bytes(), res)
		require.NoError(t, err)

		tt.mockTokenStore.AssertExpectations(t)
		tt.mockTokenNotifyStore.AssertExpectations(t)
		tt.mockCache.AssertExpectations(t)
		tt.mockJWTClient.AssertExpectations(t)
		assert.Equal(t, 200, tt.response.StatusCode())
		assert.Equal(t, 0, res.Code)
		assert.Equal(t, "success", res.Message)
	})

	t.Run("never expired", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: -1}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		tt.mockTokenStore.On("GetUserTokensByName", form.Username).Once().
			Return([]models.BcsUser{}, nil)
		tt.mockCache.On("Set", mock.Anything, mock.Anything, NeverExpiredDuration).Once().
			Return("", nil)
		tt.mockJWTClient.On("JWTSign", mock.Anything).Once().Return("", nil)
		tt.mockTokenStore.On("CreateToken", mock.Anything).Once().Return(nil)
		tt.tokenHandler.CreateToken(tt.request, tt.response)
		res := &utils.ErrorResponse{}
		err = json.Unmarshal(tt.recorder.Body.Bytes(), res)
		require.NoError(t, err)

		tt.mockTokenStore.AssertExpectations(t)
		tt.mockTokenNotifyStore.AssertExpectations(t)
		tt.mockCache.AssertExpectations(t)
		tt.mockJWTClient.AssertExpectations(t)
		assert.Equal(t, 200, tt.response.StatusCode())
		assert.Equal(t, 0, res.Code)
		assert.Equal(t, "success", res.Message)
	})
}

func TestGetToken(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 3600}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		// we can't set path parameter here, so we use empty username to check user token permission
		tt.request.SetAttribute(constant.CurrentUserAttr, &models.BcsUser{Name: "", UserType: sqlstore.PlainUser})
		tt.mockTokenStore.On("GetUserTokensByName", mock.Anything).Once().
			Return([]models.BcsUser{}, nil)
		tt.tokenHandler.GetToken(tt.request, tt.response)

		tt.mockTokenStore.AssertExpectations(t)
		data := make([]TokenResp, 0)
		assert.Equal(t, utils.CreateResponseData(nil, "success", data), tt.recorder.Body.String())
	})

	t.Run("has token", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 3600}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		// we can't set path parameter here, so we use empty username to check user token permission
		now := time.Now()
		tt.request.SetAttribute(constant.CurrentUserAttr, &models.BcsUser{Name: "", UserType: sqlstore.PlainUser})
		tt.mockTokenStore.On("GetUserTokensByName", mock.Anything).Once().
			Return([]models.BcsUser{{Name: "test", UserToken: "token", ExpiresAt: now.Add(time.Hour)}}, nil)
		tt.tokenHandler.GetToken(tt.request, tt.response)

		tt.mockTokenStore.AssertExpectations(t)
		expiredAt := now.Add(time.Hour)
		status := TokenStatusActive
		data := []TokenResp{{
			Token:     "token",
			Status:    &status,
			ExpiredAt: &expiredAt,
		}}
		assert.Equal(t, utils.CreateResponseData(nil, "success", data), tt.recorder.Body.String())
	})

	t.Run("has expired token", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 3600}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		// we can't set path parameter here, so we use empty username to check user token permission
		tt.request.SetAttribute(constant.CurrentUserAttr, &models.BcsUser{Name: "", UserType: sqlstore.PlainUser})
		tt.mockTokenStore.On("GetUserTokensByName", mock.Anything).Once().
			Return([]models.BcsUser{{Name: "test", UserToken: "token", ExpiresAt: time.Unix(1, 1)}}, nil)
		tt.tokenHandler.GetToken(tt.request, tt.response)

		tt.mockTokenStore.AssertExpectations(t)
		expiredAt := time.Unix(1, 1)
		status := TokenStatusExpired
		data := []TokenResp{{
			Token:     "token",
			Status:    &status,
			ExpiredAt: &expiredAt,
		}}
		assert.Equal(t, utils.CreateResponseData(nil, "success", data), tt.recorder.Body.String())
	})

	t.Run("never expired", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: -1}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		// we can't set path parameter here, so we use empty username to check user token permission
		now := time.Now()
		tt.request.SetAttribute(constant.CurrentUserAttr, &models.BcsUser{Name: "", UserType: sqlstore.PlainUser})
		tt.mockTokenStore.On("GetUserTokensByName", mock.Anything).Once().
			Return([]models.BcsUser{{Name: "test", UserToken: "token", ExpiresAt: now.Add(math.MaxInt64)}}, nil)
		tt.tokenHandler.GetToken(tt.request, tt.response)

		tt.mockTokenStore.AssertExpectations(t)
		status := TokenStatusActive
		data := []TokenResp{{
			Token:  "token",
			Status: &status,
		}}
		assert.Equal(t, utils.CreateResponseData(nil, "success", data), tt.recorder.Body.String())
	})
}

func TestDeleteToken(t *testing.T) {
	form := CreateTokenForm{Username: "test", Expiration: 3600}
	tt, err := newTestToken(form)
	require.NoError(t, err)

	tt.mockTokenStore.On("GetTokenByCondition", mock.Anything).Once().
		Return(&models.BcsUser{Name: form.Username})
	tt.mockCache.On("Del", mock.Anything).Once().Return(uint64(1), nil)
	tt.mockTokenStore.On("DeleteToken", mock.Anything).Once().
		Return(nil)
	tt.tokenHandler.DeleteToken(tt.request, tt.response)

	tt.mockCache.AssertExpectations(t)
	tt.mockTokenStore.AssertExpectations(t)
	assert.Equal(t, utils.CreateResponseData(nil, "success", nil), tt.recorder.Body.String())
}

func TestUpdateToken(t *testing.T) {
	t.Run("token is expired", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 1}
		token := ""
		tt, err := newTestToken(form)
		require.NoError(t, err)

		tt.mockTokenStore.On("GetTokenByCondition", mock.Anything).Once().
			Return(&models.BcsUser{Name: form.Username})
		tt.mockJWTClient.On("JWTSign", mock.Anything).Once().Return("", nil)
		tt.mockCache.On("Set", constant.TokenKeyPrefix+token, "", time.Second).Once().
			Return(mock.Anything, nil)
		tt.mockTokenStore.On("UpdateToken", mock.Anything, mock.Anything).Once().
			Return(nil)
		tt.mockTokenNotifyStore.On("DeleteTokenNotify", token).Once().Return(nil)
		tt.tokenHandler.UpdateToken(tt.request, tt.response)
		res := &utils.ErrorResponse{}
		err = json.Unmarshal(tt.recorder.Body.Bytes(), res)
		require.NoError(t, err)

		tt.mockCache.AssertExpectations(t)
		tt.mockTokenStore.AssertExpectations(t)
		tt.mockTokenNotifyStore.AssertExpectations(t)
		tt.mockJWTClient.AssertExpectations(t)
		assert.Equal(t, true, res.Result)
		assert.Equal(t, 0, res.Code)
		assert.NotNil(t, res.Data)
	})

	t.Run("refresh token success", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 1}
		token := ""
		tt, err := newTestToken(form)
		require.NoError(t, err)

		tt.mockTokenStore.On("GetTokenByCondition", mock.Anything).Once().
			Return(&models.BcsUser{Name: form.Username, ExpiresAt: time.Now().Add(time.Second)})
		tt.mockJWTClient.On("JWTSign", mock.Anything).Once().Return("", nil)
		tt.mockCache.On("Set", constant.TokenKeyPrefix+token, "", time.Second).Once().
			Return(mock.Anything, nil)
		tt.mockTokenStore.On("UpdateToken", mock.Anything, mock.Anything).Once().
			Return(nil)
		tt.mockTokenNotifyStore.On("DeleteTokenNotify", token).Once().Return(nil)
		tt.tokenHandler.UpdateToken(tt.request, tt.response)
		res := &utils.ErrorResponse{}
		err = json.Unmarshal(tt.recorder.Body.Bytes(), res)
		require.NoError(t, err)

		tt.mockCache.AssertExpectations(t)
		tt.mockTokenStore.AssertExpectations(t)
		tt.mockTokenNotifyStore.AssertExpectations(t)
		tt.mockJWTClient.AssertExpectations(t)
		assert.Equal(t, true, res.Result)
		assert.Equal(t, 0, res.Code)
		assert.NotNil(t, res.Data)
	})
}

func TestCreateTempToken(t *testing.T) {
	t.Run("invalid form", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 0}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		tt.tokenHandler.CreateTempToken(tt.request, tt.response)
		res := &utils.ErrorResponse{}
		err = json.Unmarshal(tt.recorder.Body.Bytes(), res)
		require.NoError(t, err)

		assert.Equal(t, 400, tt.response.StatusCode())
		assert.Equal(t, common.BcsErrApiBadRequest, res.Code)
	})

	t.Run("admin user create token", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 3600}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		tt.request.SetAttribute(constant.CurrentUserAttr, &models.BcsUser{Name: "admin", UserType: sqlstore.AdminUser})
		tt.mockJWTClient.On("JWTSign", mock.Anything).Once().Return("", nil)
		tt.mockCache.On("Set", mock.Anything, mock.Anything, time.Duration(form.Expiration)*time.Second).Once().
			Return("", nil)
		tt.mockTokenStore.On("CreateTemporaryToken", mock.Anything).Once().Return(nil)
		tt.tokenHandler.CreateTempToken(tt.request, tt.response)
		res := &utils.ErrorResponse{}
		err = json.Unmarshal(tt.recorder.Body.Bytes(), res)
		require.NoError(t, err)

		tt.mockTokenStore.AssertExpectations(t)
		tt.mockTokenNotifyStore.AssertExpectations(t)
		tt.mockCache.AssertExpectations(t)
		tt.mockJWTClient.AssertExpectations(t)
		assert.Equal(t, 200, tt.response.StatusCode())
		assert.Equal(t, 0, res.Code)
		assert.Equal(t, "success", res.Message)
	})

	t.Run("client user create token", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 3600}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		tt.request.SetAttribute(constant.CurrentUserAttr, &models.BcsUser{Name: "admin", UserType: sqlstore.ClientUser})
		tt.mockJWTClient.On("JWTSign", mock.Anything).Once().Return("", nil)
		tt.mockCache.On("Set", mock.Anything, mock.Anything, time.Duration(form.Expiration)*time.Second).Once().
			Return("", nil)
		tt.mockTokenStore.On("CreateTemporaryToken", mock.Anything).Once().Return(nil)
		tt.tokenHandler.CreateTempToken(tt.request, tt.response)
		res := &utils.ErrorResponse{}
		err = json.Unmarshal(tt.recorder.Body.Bytes(), res)
		require.NoError(t, err)

		tt.mockTokenStore.AssertExpectations(t)
		tt.mockTokenNotifyStore.AssertExpectations(t)
		tt.mockCache.AssertExpectations(t)
		tt.mockJWTClient.AssertExpectations(t)
		assert.Equal(t, 200, tt.response.StatusCode())
		assert.Equal(t, 0, res.Code)
		assert.Equal(t, "success", res.Message)
	})

	t.Run("not allow to access token", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 3600}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		tt.request.SetAttribute(constant.CurrentUserAttr, &models.BcsUser{Name: "user1", UserType: sqlstore.PlainUser})
		tt.tokenHandler.CreateTempToken(tt.request, tt.response)
		res := &utils.ErrorResponse{}
		err = json.Unmarshal(tt.recorder.Body.Bytes(), res)
		require.NoError(t, err)

		assert.Equal(t, 401, tt.response.StatusCode())
		assert.Equal(t, common.BcsErrApiUnauthorized, res.Code)
	})

	t.Run("can't create by self", func(t *testing.T) {
		form := CreateTokenForm{Username: "test", Expiration: 3600}
		tt, err := newTestToken(form)
		require.NoError(t, err)

		tt.request.SetAttribute(constant.CurrentUserAttr, &models.BcsUser{Name: "test", UserType: sqlstore.PlainUser})
		tt.tokenHandler.CreateTempToken(tt.request, tt.response)
		res := &utils.ErrorResponse{}
		err = json.Unmarshal(tt.recorder.Body.Bytes(), res)
		require.NoError(t, err)

		assert.Equal(t, 401, tt.response.StatusCode())
		assert.Equal(t, common.BcsErrApiUnauthorized, res.Code)
	})

}
