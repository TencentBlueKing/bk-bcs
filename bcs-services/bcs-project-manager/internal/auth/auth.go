/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package auth xxx
package auth

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	jwtGo "github.com/golang-jwt/jwt/v4"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
)

var (
	jwtClient *jwt.JWTClient
)

// SetJwtClient init jwt client
func SetJwtClient() error {
	var err error
	// 组装 jwt client
	jwtOpt, err := getJWTOpt()
	if err != nil {
		logging.Error("init jwt client failed, err:%s", err.Error())
		return errorx.NewAuthErr(fmt.Sprintf("parse jwt key error: %s", err.Error()))
	}
	if jwtClient, err = jwt.NewJWTClient(*jwtOpt); err != nil {
		logging.Error("init jwt client failed, err:%s", err.Error())
		return err
	}
	return nil
}

// GetJwtClient return jwt client
func GetJwtClient() *jwt.JWTClient {
	// 组装 jwt client
	return jwtClient
}

func getJWTOpt() (*jwt.JWTOptions, error) {
	jwtOpt := &jwt.JWTOptions{
		VerifyKeyFile: config.GlobalConf.JWT.PublicKeyFile,
		SignKeyFile:   config.GlobalConf.JWT.PrivateKeyFile,
	}
	publicKey := config.GlobalConf.JWT.PublicKey
	privateKey := config.GlobalConf.JWT.PrivateKey

	if publicKey != "" {
		key, err := jwtGo.ParseRSAPublicKeyFromPEM([]byte(publicKey))
		if err != nil {
			return nil, err
		}
		jwtOpt.VerifyKey = key
	}
	if privateKey != "" {
		key, err := jwtGo.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
		if err != nil {
			return nil, err
		}
		jwtOpt.SignKey = key
	}
	return jwtOpt, nil
}
