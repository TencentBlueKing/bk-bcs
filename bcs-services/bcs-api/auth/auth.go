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

package auth

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type BcsAuth interface {
	GetToken(header http.Header) (*Token, error)
	Allow(token *Token, action Action, resource Resource) (bool, error)
}

type Action string

const (
	ActionManage Action = "cluster-manager"
	ActionRead   Action = "cluster-readonly"

	TokenDefaultExpireTime = 2 * time.Hour
	TokenRandomLength      = 10
)

type Token struct {
	Token      string    `json:"token"`
	Username   string    `json:"username"`
	Message    string    `json:"message"`
	ExpireTime time.Time `json:"expire_time"`

	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

func (t *Token) Sign(duration time.Duration) {
	if duration == 0 {
		duration = TokenDefaultExpireTime
	}
	t.ExpireTime = time.Now().Add(duration)
}

func (t *Token) Generate() {
	t.Token = fmt.Sprintf("%d_%s", time.Now().Unix(), randomString())
}

type Resource struct {
	ClusterID string `json:"cluster_id"`
	Namespace string `json:"namespace"`
}

var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomString() string {
	b := make([]rune, TokenRandomLength)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
