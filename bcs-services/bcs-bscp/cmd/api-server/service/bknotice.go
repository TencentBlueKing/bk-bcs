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
	"net/http"

	"github.com/go-chi/render"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components/bknotice"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// bkNoticeService is http handler for bknotice service.
type bkNoticeService struct {
}

// GetCurrentAnnouncements get current announcements
func (s *bkNoticeService) GetCurrentAnnouncements(w http.ResponseWriter, r *http.Request) {
	// Prepare the new request
	lang := tools.GetLangFromReq(r)
	annotations, err := bknotice.GetCurrentAnnouncements(r.Context(), lang)
	if err != nil {
		logs.Errorf("get current announcements failed, err %s", err.Error())
	}
	_ = render.Render(w, r, rest.OKRender(annotations))
}

func newBKNoticeService() (*bkNoticeService, error) {

	service := &bkNoticeService{}

	return service, nil
}
