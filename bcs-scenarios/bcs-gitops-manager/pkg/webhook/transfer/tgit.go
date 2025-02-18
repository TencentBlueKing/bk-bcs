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

package transfer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
)

const (
	// nolint unused
	tgitTraceID = "X-TRACE-ID"
)

// Interface defines the interface of transfer webhook
type Interface interface {
	Transfer(context.Context, []byte) ([]byte, error)
}

// TGitHandler defines the tgit implementation of transfer handler
type TGitHandler struct{}

// NewTGitHandler create the handler of TGit
func NewTGitHandler() Interface {
	return &TGitHandler{}
}

// Transfer is the implementation of tgit. It used to transfer the event from tgit
// to gitlab event. Because argocd use the standard gitlab event. There only handle
//
//	the PushHook in need currently.
func (t *TGitHandler) Transfer(ctx context.Context, body []byte) ([]byte, error) {
	hookEvent := new(TGitPushHook)
	if err := json.Unmarshal(body, hookEvent); err != nil {
		return nil, errors.Wrapf(err, "unmarshal failed: %s", string(body))
	}
	blog.Infof("RequestID[%s] received '%s' webhook by user '%s': %s", ctxutils.RequestID(ctx),
		hookEvent.ObjectKind, hookEvent.UserName, hookEvent.Repository.GitHTTPURL)
	result := t.buildByPushHook(hookEvent)
	bs, err := json.Marshal(result)
	if err != nil {
		return nil, errors.Wrapf(err, "marshal failed")
	}
	return bs, nil
}

func (t *TGitHandler) transferTime(tStr string) time.Time {
	newTime, err := time.Parse("2006-01-02T15:04:05+0000", tStr)
	if err != nil {
		blog.Warnf("parse time '%s' failed: %s", tStr, err.Error())
		return time.Now()
	}
	return newTime
}
