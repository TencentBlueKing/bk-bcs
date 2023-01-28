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

package passcc

import (
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/auth"
	"github.com/patrickmn/go-cache"
)

func getPermServer() *auth.ClientAuth {
	cli := auth.NewAuthClient(auth.Options{
		Server:    "xxx",
		AppCode:   "xxx",
		AppSecret: "xxx",
	})

	return cli
}

var server = &ClientConfig{
	server:    "xxx",
	appCode:   "xxx",
	appSecret: "xxx",
	debug:     true,
	cache:     cache.New(time.Minute*5, time.Minute*60),
}

func TestConfig_GetSharedNamespaces(t *testing.T) {
	token, err := server.getAccessToken(getPermServer())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)

	namespaces, err := server.GetProjectSharedNamespaces("xxx", "xxx")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(len(namespaces))
	for _, ns := range namespaces {
		t.Logf(ns)
	}
}
