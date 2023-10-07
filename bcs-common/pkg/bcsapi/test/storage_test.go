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

package test

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
)

func TestStorageCli_QueryK8SGameDeployment(t *testing.T) { // nolint
	tlsconfig, err := ssl.ClientTslConfVerity(
		"xxx",
		"xxx",
		"xxx",
		"xxx")

	if err != nil {
		t.Errorf("ssl.ClientTslConfVerity err: %v", err)
	}

	config := &bcsapi.Config{
		Hosts:     []string{"xxx:xxx"},
		TLSConfig: tlsconfig,
		Gateway:   true,
	}

	client := bcsapi.NewClient(config)
	s := client.Storage()
	mesosNamespaces, err := s.QueryMesosNamespace("xxx")
	if err != nil {
		return
	}
	t.Logf("mesosNamespaces : %v", mesosNamespaces)

	for _, ns := range mesosNamespaces {
		t.Log(ns)
	}
}
