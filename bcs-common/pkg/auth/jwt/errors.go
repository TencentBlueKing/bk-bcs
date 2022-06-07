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

package jwt

import (
	"errors"
)

var (
	// ErrServerNotInited for server not inited error
	ErrServerNotInited = errors.New("server not init")
	// ErrJWtSignKeyEmpty for jwt signKey empty error
	ErrJWtSignKeyEmpty = errors.New("jwt options signKey empty")
	// ErrJWtVerifyKeyEmpty for jwt verify key error
	ErrJWtVerifyKeyEmpty = errors.New("jwt options verifyKey empty")
	// ErrJWtUserNameEmpty for jwt username error
	ErrJWtUserNameEmpty = errors.New("jwt uerInfo userName empty")
	// ErrJWtClientNameEmpty for jwt clientname error
	ErrJWtClientNameEmpty = errors.New("jwt uerInfo clientName empty")
	// ErrJWtSubType for jwt user type error
	ErrJWtSubType = errors.New("jwt subType err: user or client")
	// ErrTokenIsNil parse with claim return token is nil
	ErrTokenIsNil = errors.New("parse with claims return token is nil")
)
