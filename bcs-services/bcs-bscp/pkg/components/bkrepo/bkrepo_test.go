/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package bkrepo

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"bscp.io/pkg/cc"
)

func TestUpload(t *testing.T) {
	cc.InitService(cc.APIServerName)
	configFiles := strings.Split(os.Getenv("CONFIG_PATH"), ",")
	err := cc.LoadSettings(&cc.SysOption{ConfigFiles: configFiles})
	assert.NoError(t, err)

	payload, err := os.ReadFile(os.Getenv("OBJ_PATH"))
	assert.NoError(t, err)

	hash := sha256.New()
	hash.Write(payload)
	h := fmt.Sprintf("%x", hash.Sum(nil))

	raw := &http.Request{Body: ioutil.NopCloser(bytes.NewReader(payload))}
	result, err := Upload(context.Background(), raw, 2, "dummyUser", h)

	assert.NoError(t, err)
	assert.True(t, result != "")
}
