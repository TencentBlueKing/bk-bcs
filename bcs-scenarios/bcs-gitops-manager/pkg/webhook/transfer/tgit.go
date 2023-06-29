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

package transfer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"go-micro.dev/v4/metadata"
	"gopkg.in/go-playground/webhooks.v5/gitlab"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

const (
	tgitTraceID = "X-TRACE-ID"
)

// TGitHandler defines the tgit implementation of transfer handler
type TGitHandler struct{}

// NewTGitHandler create the handler of TGit
func NewTGitHandler() Interface {
	return &TGitHandler{}
}

func (t *TGitHandler) getTraceID(ctx context.Context) string {
	var traceID string
	md, ok := metadata.FromContext(ctx)
	if ok {
		traceID, ok = md.Get(tgitTraceID)
		if !ok {
			blog.Warnf("tgit not have header '%s'", tgitTraceID)
		}
	}
	if traceID == "" {
		u, err := uuid.NewV4()
		if err != nil {
			blog.Warnf("tgit generate uuid failed: %s", err.Error())
		} else {
			traceID = u.String()
		}
	}
	return traceID
}

// Transfer is the implementation of tgit. It used to transfer the event from tgit
// to gitlab event. Because argocd use the standard gitlab event. There only handle
//  the PushHook in need currently.
func (t *TGitHandler) Transfer(ctx context.Context, body []byte) ([]byte, error) {
	hookEvent := new(TGitPushHook)
	if err := json.Unmarshal(body, hookEvent); err != nil {
		return nil, errors.Wrapf(err, "unmarshal failed: %s", string(body))
	}
	traceID := t.getTraceID(ctx)
	blog.Infof("tgit[%s'] received '%s' webhook by user '%s'", traceID,
		hookEvent.ObjectKind, hookEvent.UserName)
	for i := range hookEvent.Commits {
		commit := hookEvent.Commits[i]
		blog.Infof("tgit[%s] commit '%s' author '%s': add[%v], modified[%v], removed[%v]",
			traceID, commit.ID, commit.Author, commit.Added, commit.Modified, commit.Removed)
	}
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
		blog.Warnf("[TGIT] parse time '%s' failed: %s", tStr, err.Error())
		return time.Now()
	}
	return newTime
}

// buildByPushHook will transfer tgit push webhook event to gitlab event
func (t *TGitHandler) buildByPushHook(hook *TGitPushHook) gitlab.PushEventPayload {
	commits := make([]gitlab.Commit, 0, len(hook.Commits))
	for i := range hook.Commits {
		c := hook.Commits[i]
		commits = append(commits, gitlab.Commit{
			ID:        c.ID,
			Message:   c.Message,
			Timestamp: struct{ time.Time }{t.transferTime(c.Timestamp)},
			URL:       c.URL,
			Author: gitlab.Author{
				Name:  c.Author.Name,
				Email: c.Author.Email,
			},
			Added:    c.Added,
			Modified: c.Modified,
			Removed:  c.Removed,
		})
	}
	return gitlab.PushEventPayload{
		ObjectKind:   hook.ObjectKind,
		Before:       hook.Before,
		After:        hook.After,
		Ref:          hook.Ref,
		CheckoutSHA:  hook.CheckoutSha,
		UserID:       hook.UserID,
		UserName:     hook.UserName,
		UserUsername: hook.UserName,
		UserEmail:    hook.UserEmail,
		ProjectID:    hook.ProjectID,
		Project: gitlab.Project{
			ID:              hook.ProjectID,
			Name:            hook.Repository.Name,
			Description:     hook.Repository.Description,
			WebURL:          hook.Repository.Homepage,
			GitSSSHURL:      hook.Repository.GitSSHURL,
			GitHTTPURL:      hook.Repository.GitHTTPURL,
			VisibilityLevel: hook.Repository.VisibilityLevel,
			DefaultBranch:   "master",
			Homepage:        hook.Repository.Homepage,
			URL:             hook.Repository.URL,
			SSHURL:          hook.Repository.GitSSHURL,
			HTTPURL:         hook.Repository.GitHTTPURL,
		},
		Repository: gitlab.Repository{
			Name:        hook.Repository.Name,
			URL:         hook.Repository.URL,
			Description: hook.Repository.Description,
			Homepage:    hook.Repository.Homepage,
		},
		Commits:           commits,
		TotalCommitsCount: hook.TotalCommitsCount,
	}
}

// TGitPushHook defines the hook event struct of tgit
type TGitPushHook struct {
	ObjectKind    string      `json:"object_kind"`
	OperationKind string      `json:"operation_kind"`
	ActionKind    string      `json:"action_kind"`
	Before        string      `json:"before"`
	After         string      `json:"after"`
	Ref           string      `json:"ref"`
	CheckoutSha   string      `json:"checkout_sha"`
	StartPoint    interface{} `json:"start_point"`
	UserName      string      `json:"user_name"`
	UserID        int64       `json:"user_id"`
	UserEmail     string      `json:"user_email"`
	ProjectID     int64       `json:"project_id"`
	Repository    struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		Homepage        string `json:"homepage"`
		GitHTTPURL      string `json:"git_http_url"`
		GitSSHURL       string `json:"git_ssh_url"`
		URL             string `json:"url"`
		VisibilityLevel int64  `json:"visibility_level"`
	} `json:"repository"`
	Commits []struct {
		ID              string `json:"id"`
		Message         string `json:"message"`
		Timestamp       string `json:"timestamp"`
		AuthorTimestamp string `json:"author_timestamp"`
		URL             string `json:"url"`
		Author          struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
		Added    []string `json:"added"`
		Modified []string `json:"modified"`
		Removed  []string `json:"removed"`
	} `json:"commits"`
	DiffFiles []struct {
		NewPath     string `json:"new_path"`
		OldPath     string `json:"old_path"`
		AMode       int    `json:"a_mode"`
		BMode       int    `json:"b_mode"`
		NewFile     bool   `json:"new_file"`
		RenamedFile bool   `json:"renamed_file"`
		DeletedFile bool   `json:"deleted_file"`
	} `json:"diff_files"`
	PushOptions struct {
	} `json:"push_options"`
	PushTimestamp     string      `json:"push_timestamp"`
	TotalCommitsCount int64       `json:"total_commits_count"`
	CreateAndUpdate   interface{} `json:"create_and_update"`
}
