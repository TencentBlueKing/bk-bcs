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

package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"go-micro.dev/v4/metadata"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pb "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/proto"
)

// TGitWebhook defines the webhook handler of tgit, there will transfer tgit webhook
// to gitlab webhook
func (s *Server) TGitWebhook(ctx context.Context, req *pb.TGitWebhookRequest, resp *pb.TGitWebhookResponse) error {
	blog.Infof("tgit received webhook")
	bs, err := json.Marshal(req.Body)
	if err != nil {
		blog.Errorf("tgit marshal request body failed: %s", err.Error())
		return err
	}
	blog.V(5).Infof("received tgit webhook: %s", string(bs))

	var result []byte
	result, err = s.tgitHandler.Transfer(ctx, bs)
	if err != nil {
		blog.Errorf("tagit handle transfer failed with body '%s': %s", string(bs), err.Error())
	}
	var respBody []byte
	respBody, err = s.sendToGitops(ctx, result)
	if err != nil {
		blog.Errorf("tgit send to gitops with body '%s' failed: %s", string(result), err.Error())
		return err
	}
	blog.V(5).Infof("tgit webhook response: %s", string(respBody))
	return nil
}

// GeneralWebhook defines the handler of general webhook, it will add the authorization header
func (s *Server) GeneralWebhook(ctx context.Context, req *pb.GeneralWebhookRequest,
	resp *pb.GeneralWebhookResponse) error {
	blog.Infof("general received webhook")
	bs, err := json.Marshal(req.Body)
	if err != nil {
		blog.Errorf("general marshal request body failed: %s", err.Error())
		return err
	}
	blog.V(5).Infof("received general webhook: %s", string(bs))

	var respBody []byte
	respBody, err = s.sendToGitops(ctx, bs)
	if err != nil {
		blog.Errorf("general send to gitops with body '%s' failed: %s", string(bs), err.Error())
		return err
	}
	blog.V(5).Infof("general webhook response: %s", string(respBody))
	return nil
}

func (s *Server) sendToGitops(ctx context.Context, body []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, s.op.GitOpsWebhook, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Wrapf(err, "create http request failed")
	}

	md, ok := metadata.FromContext(ctx)
	if !ok {
		blog.Warnf("parse request header from context failed")
	} else {
		for k, v := range md {
			req.Header.Set(k, v)
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.op.GitOpsToken)

	httpClient := http.DefaultClient
	var resp *http.Response
	if resp, err = httpClient.Do(req); err != nil {
		return nil, errors.Wrapf(err, "http request failed")
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read response body failed")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(errors.Errorf(string(bs)), "http response status not 200 but %d",
			resp.StatusCode)
	}
	return bs, nil
}
