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
	"io/ioutil"
	"os"
	"testing"

	"github.com/dgrijalva/jwt-go"
)

var username = "admin"

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
	// generate jwt token
	token, err := cli.JWTSign(&UserInfo{
		SubType:     User.String(),
		UserName:    username,
		ExpiredTime: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	user, err := cli.JWTDecode(token)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", user)
}

func newErrPubKeyClient() (*JWTClient, error) {
	path, _ := os.Getwd()
	privateKeyPath := path + "/jwt_file/app.rsa"
	// read
	privateKeyByte, _ := ioutil.ReadFile(privateKeyPath)

	// error public key
	publicKeyStr := `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCozhk0uWnt0FYtaDyZR1A/ejKM 
Caj8B/axFyHzrW8GV07zOwlesbXykS8OOOtJ4AO61AdhoIPAz9p08TvBFd4R2tYe 
MpCm9MeZnMJcEilFbh980JfdDMjioVdzpgJMJrnF99wYjNZpmBsPMFdBOq4K8WJL 
E4g1rOJKpJfc30YsfQIDAQAa 
-----END PUBLIC KEY-----`

	publicKey, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyStr))
	privateKey, _ := jwt.ParseRSAPrivateKeyFromPEM(privateKeyByte)

	opts := JWTOptions{
		VerifyKey: publicKey,
		SignKey:   privateKey,
	}
	cli, err := NewJWTClient(opts)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func TestErrPubKey(t *testing.T) {
	cli, err := newErrPubKeyClient()
	if err != nil {
		t.Fatal(err)
	}
	token, err := cli.JWTSign(&UserInfo{
		SubType:     User.String(),
		UserName:    username,
		ExpiredTime: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cli.JWTDecode(token)
	if err == nil {
		t.Fatal("err is not nil when error public key")
	}
}
