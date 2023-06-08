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

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"bscp.io/pkg/tools"
)

// dd if=/dev/urandom of=/tmp/10Mib.bin bs=1M count=10
func upload(ctx context.Context, host string, bizID, appID string, fileContentID string, body io.Reader) (*http.Response, error) {
	u := fmt.Sprintf("http://%s/api/v1/api/create/content/upload/biz_id/%s/app_id/%s", host, bizID, appID)
	req, err := http.NewRequestWithContext(ctx, "PUT", u, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Bkapi-File-Content-Id", fileContentID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("%d != 200", resp.StatusCode)
		}
		return nil, fmt.Errorf("%d != 200, body: %s", resp.StatusCode, body)
	}
	return resp, nil
}

func main() {
	filename := os.Getenv("filename")
	host := os.Getenv("host")
	if host == "" {
		host = "localhost:8080"
	}
	bizID := os.Getenv("biz_id")
	appID := os.Getenv("app_id")

	wg := &sync.WaitGroup{}

	d, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	fileContentID := tools.ByteSHA256(d)

	c := 20
	for i := 0; i < c; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for {
				resp, err := upload(context.Background(), host, bizID, appID, fileContentID, bytes.NewReader(d))
				if err != nil {
					fmt.Println(idx, err, resp)
				}
				time.Sleep(time.Millisecond * 100)
			}

		}(i)
	}
	wg.Wait()
}
