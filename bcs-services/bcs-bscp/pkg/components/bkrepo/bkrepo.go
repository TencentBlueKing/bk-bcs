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
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/components"
	"bscp.io/pkg/thirdparty/repo"
)

// Upload 上传文件
func Upload(ctx context.Context, raw *http.Request, bizID uint32, appId, sha256 string) (string, error) {
	config := cc.ApiServer().Repo
	opt := &repo.NodeOption{
		Project: config.BkRepo.Project,
		BizID:   bizID,
		Sign:    sha256,
	}
	url, err := repo.GenNodePath(opt)
	if err != nil {
		return "", err
	}
	endpoint, err := config.OneEndpoint()
	if err != nil {
		return "", err
	}
	u := fmt.Sprintf("%s%s", endpoint, url)

	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetHeader("Authorization", "Platform "+config.BkRepo.Token).
		SetHeader(repo.HeaderKeyUID, config.BkRepo.User).
		SetHeader(repo.HeaderKeySHA256, sha256).
		SetHeader(repo.HeaderKeyOverwrite, "true").
		SetBody(raw.Body).
		Put(u)

	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return "", errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	return string(resp.Body()), nil
}
