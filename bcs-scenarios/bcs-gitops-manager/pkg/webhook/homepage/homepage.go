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

// Package homepage used to recorder the commit message
package homepage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	"gopkg.in/go-playground/webhooks.v5/github"
	"gopkg.in/go-playground/webhooks.v5/gitlab"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
)

// Recorder defines the recorder of home page
type Recorder struct {
	github *github.Webhook
	gitlab *gitlab.Webhook
}

// NewRecorder create the recorder instance
func NewRecorder() (*Recorder, error) {
	githubHandler, err := github.New()
	if err != nil {
		return nil, fmt.Errorf("unable to init GitHub webhook: %v", err)
	}
	gitlabHandler, err := gitlab.New()
	if err != nil {
		return nil, fmt.Errorf("unable to init GitLab webhook: %v", err)
	}
	return &Recorder{
		github: githubHandler,
		gitlab: gitlabHandler,
	}, nil
}

// RecordEvent will record event by request body
func (r *Recorder) RecordEvent(ctx context.Context, req *http.Request) error {
	var payload interface{}
	var err error
	switch {
	case req.Header.Get("X-GitHub-Event") != "":
		payload, err = r.github.Parse(req, github.PushEvent, github.PullRequestEvent, github.PingEvent)
	case req.Header.Get("X-Gitlab-Event") != "":
		payload, err = r.gitlab.Parse(req, gitlab.PushEvents, gitlab.TagEvents, gitlab.MergeRequestEvents)
	default:
		return fmt.Errorf("unknown webhook event")
	}
	if err != nil {
		return errors.Wrapf(err, "create parse webhook event failed")
	}
	switch p := payload.(type) {
	case github.PushPayload:
		for _, commit := range p.Commits {
			hpInfo := &homePageInfo{
				BcsUsername: commit.Author.Username,
				Email:       commit.Author.Email,
				CommitID:    commit.ID,
				Url:         commit.URL,
				Modified:    commit.Modified,
				Added:       commit.Added,
				Removed:     commit.Removed,
			}
			r.PrintHomePageInfo(ctx, hpInfo)
		}
	case gitlab.PushEventPayload:
		for _, commit := range p.Commits {
			hpInfo := &homePageInfo{
				BcsUsername: commit.Author.Name,
				Email:       commit.Author.Email,
				CommitID:    commit.ID,
				Url:         commit.URL,
				Modified:    commit.Modified,
				Added:       commit.Added,
				Removed:     commit.Removed,
			}
			r.PrintHomePageInfo(ctx, hpInfo)
		}
	}
	return nil
}

// PrintHomePageInfo will print commit info to stdout
func (r *Recorder) PrintHomePageInfo(ctx context.Context, hpInfo *homePageInfo) {
	if len(hpInfo.Added)+len(hpInfo.Removed)+len(hpInfo.Modified) == 0 {
		blog.Warnf("RequestID[%s] repo '%s' commit '%s' with user '%s' not changed files",
			middleware.RequestID(ctx), hpInfo.Url, hpInfo.CommitID, hpInfo.BcsUsername)
		return
	}
	bs, err := hpInfo.Build()
	if err != nil {
		blog.Errorf("RequestID[%s] build homepage info for '%s' failed: %s",
			middleware.RequestID(ctx), hpInfo.Url, err.Error())
	} else {
		blog.Infof("%s", string(bs))
	}
}

// homePageInfo defines the info of homepage
type homePageInfo struct {
	BcsUsername string   `json:"bcs_username"`
	Email       string   `json:"email"`
	CommitID    string   `json:"commit_id"`
	Url         string   `json:"url"`
	Modified    []string `json:"modified"`
	Added       []string `json:"added"`
	Removed     []string `json:"removed"`
}

// Build return the marshal or error
func (inf *homePageInfo) Build() ([]byte, error) {
	if inf.BcsUsername == "" {
		if !strings.HasSuffix(inf.Email, "@tencent.com") {
			return nil, fmt.Errorf("not have auth info: %s/%s", inf.BcsUsername, inf.Email)
		}
		inf.BcsUsername = strings.TrimSuffix(inf.Email, "@tencent.com")
	}
	bs, err := json.Marshal(inf)
	if err != nil {
		return nil, errors.Wrapf(err, "marshal homepage info failed")
	}
	return bs, nil
}
