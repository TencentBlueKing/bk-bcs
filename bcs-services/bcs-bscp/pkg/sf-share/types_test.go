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

package sfs

import (
	"reflect"
	"strings"
	"testing"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
)

func TestTLSBytes(t *testing.T) {
	caFile := "caBody"
	certFile := "certBody"
	keyFile := "keyBody"

	tls := TLSBytes{
		CaFileBytes:   caFile,
		CertFileBytes: certFile,
		KeyFileBytes:  keyFile,
	}

	bytes, err := jsoni.Marshal(tls)
	if err != nil {
		t.Errorf("encode tls failed, err: %v", err)
		return
	}

	encoded := string(bytes)
	if strings.Contains(encoded, caFile) && strings.Contains(encoded, certFile) && strings.Contains(encoded, keyFile) {
		t.Errorf("encoded with base64 failed")
		return
	}

	decoded := new(TLSBytes)
	if err := jsoni.Unmarshal(bytes, decoded); err != nil {
		t.Errorf("deocde tls failed, err: %v", err)
		return
	}

	if !reflect.DeepEqual(tls, *decoded) {
		t.Errorf("decoded tls is not what we expected!")
		return
	}

}
