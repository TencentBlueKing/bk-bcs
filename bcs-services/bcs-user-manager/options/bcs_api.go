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

package options

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	netUrl "net/url"

	"github.com/pkg/errors"
)

// BcsAPI bcs api config
type BcsAPI struct {
	Host  string `json:"host" usage:"enable http host"`
	Token string `json:"token" usage:"token for calling service"`
}

// HttpRequest : http request
func (bcsAPI *BcsAPI) HttpRequest(method, url string, request, response interface{}) (err error) {
	// parse config host
	_, err = netUrl.Parse(bcsAPI.Host)
	if err != nil {
		return fmt.Errorf("url failed %v", err)
	}
	url = bcsAPI.Host + url
	// http client
	client := &http.Client{}
	byteRequest, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// organizational http parameters
	req, err := http.NewRequest(method, url, bytes.NewReader(byteRequest))
	if err != nil {
		return err
	}

	// add auth token
	if len(bcsAPI.Token) != 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bcsAPI.Token))
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// defer close the resp body
	defer func() {
		err2 := resp.Body.Close()
		if err2 != nil && err != nil {
			err = fmt.Errorf("http request error:%s and http close error: %s", err, err2)
			return
		}

		if err2 != nil {
			err = err2
			return
		}
	}()

	// Non 200 status returned error
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("http response status not 200 but %d",
			resp.StatusCode)
	}

	// response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Unmarshal into response
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return err
	}

	return nil
}
