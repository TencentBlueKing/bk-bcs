/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package authorization

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	pbauthserver "bk-bscp/internal/protocol/authserver"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/kit"
)

// Authorize handles auth check action.
func Authorize(kit kit.Kit, object, action string,
	authSvrCli pbauthserver.AuthClient, timeout time.Duration) (bool, error) {

	isAuthOpen, err := validate(kit, object, action, authSvrCli)
	if err != nil {
		return false, fmt.Errorf("can't authorize, %+v", err)
	}
	if !isAuthOpen {
		// authorization is not open, just pass it.
		return true, nil
	}
	return authorize(kit, object, action, authSvrCli, timeout)
}

func validate(kit kit.Kit, object, action string, authSvrCli pbauthserver.AuthClient) (bool, error) {
	if kit.Ctx == nil {
		return false, errors.New("empty ctx")
	}
	if len(kit.Rid) == 0 {
		return false, errors.New("empty rid")
	}
	if len(kit.User) == 0 {
		return false, errors.New("empty user")
	}
	if len(kit.Authorization) == 0 {
		return false, errors.New("empty authorization flag")
	}

	if len(object) == 0 || len(action) == 0 {
		return false, fmt.Errorf("empty auth metadata, object[%+v] action[%+v]", object, action)
	}

	if authSvrCli == nil {
		return false, errors.New("empty auth server client")
	}

	isAuthOpen, err := strconv.ParseBool(kit.Authorization)
	if err != nil {
		return false, fmt.Errorf("invalid authorization key[%s], %+v", kit.Authorization, err)
	}
	return isAuthOpen, nil
}

func authorize(kit kit.Kit, object, action string,
	authSvrCli pbauthserver.AuthClient, timeout time.Duration) (bool, error) {

	req := &pbauthserver.AuthorizeReq{
		Seq:      kit.Rid,
		Metadata: &pbauthserver.AuthMetadata{V0: kit.User, V1: object, V2: action},
	}

	ctx, cancel := context.WithTimeout(kit.Ctx, timeout)
	defer cancel()

	resp, err := authSvrCli.Authorize(ctx, req)
	if err != nil {
		return false, err
	}

	if resp.Code == pbcommon.ErrCode_E_AUTH_NOT_AUTHORIZED {
		// action not authorized.
		return false, nil
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return false, errors.New(resp.Message)
	}

	return true, nil
}
