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

package service

import (
	"fmt"
	"io"
	"net/http"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/config"
)

// bkNoticeService is http handler for bknotice service.
type bkNoticeService struct {
}

// GetCurrentAnnouncements get current announcements
func (s *bkNoticeService) GetCurrentAnnouncements(w http.ResponseWriter, r *http.Request) {
	// Prepare the new request

	proxyURL := fmt.Sprintf("%s/v1/announcement/get_current_announcements/?platform=%s",
		cc.ApiServer().BKNotice.Host, cc.ApiServer().Esb.AppCode)

	proxyReq, err := http.NewRequest("GET", proxyURL, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authHeader := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\"}",
		cc.ApiServer().Esb.AppCode, cc.ApiServer().Esb.AppCode)

	proxyReq.Header.Set("X-Bkapi-Authorization", authHeader)

	// Send the request to the target API
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy the response headers and status code
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func newBKNoticeService() (*bkNoticeService, error) {

	service := &bkNoticeService{}

	return service, nil
}
