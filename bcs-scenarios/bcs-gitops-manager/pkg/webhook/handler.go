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

package webhook

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	"go-micro.dev/v4/metadata"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	pb "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/proto"
)

// TGitWebhook defines the webhook handler of tgit, there will transfer tgit webhook
// to gitlab webhook
func (s *Server) TGitWebhook(ctx context.Context, req *pb.TGitWebhookRequest, resp *pb.TGitWebhookResponse) error {
	_, span := s.tracer.Start(ctx, "tgit")
	defer span.End()

	blog.Infof("RequestID[%s] tgit received webhook", middleware.RequestID(ctx))
	result, err := s.tgitHandler.Transfer(ctx, req.Data)
	if err != nil {
		blog.Errorf("RequestID[%s] tagit handle transfer failed with body '%s': %s",
			middleware.RequestID(ctx), string(req.Data), err.Error())
		return err
	}
	var respBody []byte
	respBody, err = s.sendToGitops(ctx, result)
	if err != nil {
		blog.Errorf("RequestID[%s] tgit send to gitops with body '%s' failed: %s",
			middleware.RequestID(ctx), string(result), err.Error())
		return err
	}
	blog.V(5).Infof("RequestID[%s] tgit webhook response: %s", middleware.RequestID(ctx), string(respBody))
	return nil
}

// GeneralWebhook defines the handler of general webhook, it will add the authorization header
func (s *Server) GeneralWebhook(
	ctx context.Context, req *pb.GeneralWebhookRequest, resp *pb.GeneralWebhookResponse) error {
	_, span := s.tracer.Start(ctx, "general")
	defer span.End()

	blog.Infof("RequestID[%s] general received webhook", middleware.RequestID(ctx))
	respBody, err := s.sendToGitops(ctx, req.Data)
	if err != nil {
		blog.Errorf("RequestID[%s] general send to gitops with body '%s' failed: %s",
			middleware.RequestID(ctx), string(req.Data), err.Error())
		return err
	}
	blog.V(5).Infof("RequestID[%s] general webhook response: %s", middleware.RequestID(ctx), string(respBody))
	return nil
}

func (s *Server) createWebhookRequest(ctx context.Context, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, s.op.GitOpsWebhook, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Wrapf(err, "create http request failed")
	}
	md, ok := metadata.FromContext(ctx)
	if !ok {
		blog.Warnf("parse request header from context failed")
	} else {
		for k, v := range md {
			if k == ":Authority" {
				continue
			}
			req.Header.Add(k, v)
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.op.GitOpsToken)
	return req, nil
}

func (s *Server) sendToGitops(ctx context.Context, body []byte) ([]byte, error) {
	recordReq, err := s.createWebhookRequest(ctx, body)
	if err != nil {
		return nil, errors.Wrapf(err, "create webhook request for record failed")
	}
	if err = s.recorder.RecordEvent(ctx, recordReq); err != nil {
		return nil, errors.Wrapf(err, "record event failed")
	}

	req, err := s.createWebhookRequest(ctx, body)
	if err != nil {
		return nil, errors.Wrapf(err, "create webhook request failed")
	}
	var resp *http.Response
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return nil, errors.Wrapf(err, "http request failed")
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read response body failed")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(errors.Errorf(string(bs)), "http response status not 200 but %d",
			resp.StatusCode)
	}
	return bs, nil
}
