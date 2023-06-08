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

package apiserver

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// dd if=/dev/urandom of=/tmp/10Mib.bin bs=1M count=10
func upload(ctx context.Context, filename, host string, bizID, appID string) (*http.Response, error) {
	d, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf("http://%s/api/v1/api/create/content/upload/biz_id/%s/app_id/%s", host, bizID, appID)
	req, err := http.NewRequestWithContext(ctx, "PUT", u, bytes.NewReader(d))
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%d != 200", resp.StatusCode)
	}
	return resp, nil
}

func TestUpload(t *testing.T) {
	filename := os.Getenv("filename")
	host := os.Getenv("host")
	if host == "" {
		host = "localhost:8080"
	}
	bizID := os.Getenv("biz_id")
	appID := os.Getenv("app_id")
	resp, err := upload(context.Background(), filename, host, bizID, appID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDownload(t *testing.T) {

}
