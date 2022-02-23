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
	"os"
	"testing"
)

// openssl genrsa -out app.rsa -aes256 123456
// openssl rsa -in app.rsa -pubout > app.rsa.pub

func NewClient() (*JWTClient, error) {
	path, _ := os.Getwd()

	opts := JWTOptions{
		VerifyKeyFile: path + "/jwt_file/app.rsa.pub",
		SignKeyFile:   path + "/jwt_file/app.rsa",
	}
	cli, err := NewJWTClient(opts)
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func TestJWTClient_JWTSign(t *testing.T) {
	cli, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	token, err := cli.JWTSign(&UserInfo{
		SubType:     User.String(),
		UserName:    "james",
		ExpiredTime: 100,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(token)
}

func TestJWTClient_JWTDecode(t *testing.T) {
	cli, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	jwtTokenString := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWJfdHlwZSI6InVzZXIiLCJ1c2VybmFtZSI6ImphbWVzIiwiY2xpZW50X2lkIjoiIiwiY2xpZW50X3NlY3JldCI6IiIsImV4cCI6MTY0MzAwNjkzMywiaXNzIjoiQkNTIn0.GW_iX7a8AfVKu7tuWrBDemc3J7GWbWZDVh4H_HerJCSvKuJA48PwAn_QMzw5V2YgkZMg6_kiSuhbGWwbsWfnUnT2880kA-hB01duIbU8j8fqsnouzb1-Srz7pY4_bkxNpXJPOkvW7ydY3C1Up-PseU-TdUCAgyJxnn8DsouUU6s"
	user, err := cli.JWTDecode(jwtTokenString)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", user)
}
