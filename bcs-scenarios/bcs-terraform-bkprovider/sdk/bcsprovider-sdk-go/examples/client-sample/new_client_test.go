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

// Package clientSample 测试
package clientSample

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/utils"
)

func TestNewClient(t *testing.T) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			// NOCC:gas/tls(设计如此)
			InsecureSkipVerify: true, // nolint
		},
	}

	client := &http.Client{
		Transport: transport,
	}

	resp, err := client.Do(nil)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	log.Printf("body: %s", string(bs))
}

func TestNewSdkClient(t *testing.T) {
	config := &options.Config{
		Username:       "huiwen",
		Token:          "xxxx",
		BcsGatewayAddr: "http://xxxx",
	}

	client, err := sdk.NewClient(config)
	if err != nil {
		panic(fmt.Sprintf("err: %s", err.Error()))
	}

	log.Printf("config: %s", utils.ObjToPrettyJson(client.Config()))
}
