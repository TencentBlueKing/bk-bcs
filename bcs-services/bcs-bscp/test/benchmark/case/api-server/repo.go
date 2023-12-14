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

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"k8s.io/klog/v2"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// dd if=/dev/urandom of=/tmp/100Mib.bin bs=1M count=100
func upload(ctx context.Context, host string, bizID, appID string, sign string, body io.Reader) (*http.Response, error) {
	u := fmt.Sprintf("http://%s/api/v1/api/create/content/upload/biz_id/%s/app_id/%s", host, bizID, appID)
	req, err := http.NewRequestWithContext(ctx, "PUT", u, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Bkapi-File-Content-Id", sign)
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

func download(ctx context.Context, host string, bizID, appID string, sign string, body io.Reader) (*http.Response, error) {
	u := fmt.Sprintf("http://%s/api/v1/api/get/content/download/biz_id/%s/app_id/%s", host, bizID, appID)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Bkapi-File-Content-Id", sign)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

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
	if filename == "" {
		filename = "/tmp/100Mib.bin"
	}
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
	sign := tools.ByteSHA256(d)
	klog.InfoS("file", "id", sign)

	c := 1
	for i := 0; i < c; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for {
				st := time.Now()
				resp, err := upload(context.Background(), host, bizID, appID, sign, bytes.NewReader(d))
				// resp, err := download(context.Background(), host, bizID, appID, sign, bytes.NewReader(d))
				if err != nil {
					klog.ErrorS(err, "idx", idx, "resp", resp)
					continue
				}

				hasher := sha256.New()
				io.Copy(hasher, resp.Body)
				klog.InfoS("resp", "idx", idx, "duration", time.Since(st), "id", fmt.Sprintf("%x", hasher.Sum(nil)))
				resp.Body.Close()

				time.Sleep(time.Millisecond * 100)
			}

		}(i)
	}
	wg.Wait()
}
