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

package git

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/pkg/apis/requester"
	precheck "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/proto"
)

type tGitClient struct {
	httpCli requester.Requester
	opts    *clientOpts
}

// NewTGit new tGit client
func NewTGit(opts *clientOpts) Client {
	return &tGitClient{opts: opts, httpCli: requester.NewRequester()}
}

const (
	getBranchDetailPath = "/api/v3/projects/%s/repository/branches/%s?private_token=%s"
	getMrInfoPath       = "/api/v3/projects/%s/merge_request/iid/%s?private_token=%s"
	commentMrPath       = "/api/v3/projects/%s/merge_request/%d/reviewer/summary?private_token=%s"
	submitCheckState    = "/api/v3/projects/%s/commit/%s/statuses?private_token=%s"
	getTagPath          = "/api/v3/projects/%s/repository/tags/%s?private_token=%s"
)

// GetMrInfo get mr info
func (c *tGitClient) GetMrInfo(ctx context.Context, repoAddr, token, mrIID string) (*precheck.MRInfoData, error) {
	projectFullPath := enCodeRepoFullpath(repoAddr)
	uri := fmt.Sprintf("%s%s", c.opts.endpoint, fmt.Sprintf(getMrInfoPath, projectFullPath, mrIID, token))
	rsp, err := c.httpCli.DoGetRequest(uri, nil)
	if err != nil {
		return nil, fmt.Errorf("request git api error:%s", err.Error())
	}

	mrInfoRes := &tGitMrInfoResp{}
	err = json.Unmarshal(rsp, mrInfoRes)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal rsp failed:%s, rsp:%s", err.Error(), string(rsp))
	}

	mrInfo := &precheck.MRInfoData{
		SourceBranch: mrInfoRes.SourceBranch,
		TargetBranch: mrInfoRes.TargetBranch,
		SourceCommit: mrInfoRes.SourceCommit,
		TargetCommit: mrInfoRes.TargetCommit,
		Creator:      mrInfoRes.Author.Username,
		CreateTime:   mrInfoRes.CreateTime,
		UpdateTime:   mrInfoRes.UpdateTime,
		Title:        mrInfoRes.Title,
		MrMessage:    mrInfoRes.Description,
		Repository:   repoAddr,
		Id:           uint32(mrInfoRes.ID),
		Iid:          uint32(mrInfoRes.IID),
	}
	return mrInfo, nil
}

// CommentMR comment
func (c *tGitClient) CommentMR(ctx context.Context, repoAddr, token, mrIID, comment string) error {
	mrInfo, err := c.GetMrInfo(ctx, repoAddr, token, mrIID)
	if err != nil {
		return err
	}
	projectFullPath := enCodeRepoFullpath(repoAddr)
	uri := fmt.Sprintf("%s%s", c.opts.endpoint, fmt.Sprintf(commentMrPath, projectFullPath, mrInfo.Id, token))
	// blog.Infof("commentMR url:%s", uri)
	req := map[string]string{
		"reviewer_event": "comment",
		"summary":        comment,
	}
	reqStr, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal req failed:%s", err.Error())
	}
	rsp, err := c.httpCli.DoPutRequest(uri, nil, reqStr)
	if err != nil {
		return fmt.Errorf("request git api error:%s", err.Error())
	}

	tGitResp := &tGitMrCommentResp{}
	err = json.Unmarshal(rsp, tGitResp)
	if err != nil {
		return fmt.Errorf("json unmarshal rsp failed:%s, rsp:%s", err.Error(), string(rsp))
	}
	return nil
}

// SubmitCheckState submit state
func (c *tGitClient) SubmitCheckState(ctx context.Context, checkSystem, state, targetUrl, description,
	token string, block bool, req *precheck.PreCheckTask) error {
	projectFullPath := enCodeRepoFullpath(req.RepositoryAddr)
	uri := fmt.Sprintf("%s%s", c.opts.endpoint,
		fmt.Sprintf(submitCheckState, projectFullPath, req.CheckRevision, token))
	// blog.Infof("SubmitCheckState url:%s", uri)
	body := map[string]interface{}{
		"target_url":      targetUrl,
		"description":     description,
		"block":           block,
		"state":           state,
		"context":         checkSystem,
		"target_branches": make([]string, 0),
	}
	reqStr, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal req failed:%s", err.Error())
	}
	_, err = c.httpCli.DoPostRequest(uri, nil, reqStr)
	if err != nil {
		return fmt.Errorf("request git api error:%s", err.Error())
	}
	return nil
}

func (c *tGitClient) GetBranchDetail(ctx context.Context, repoAddr, branch, token string) (*TGitBranch, error) {
	projectFullPath := enCodeRepoFullpath(repoAddr)
	uri := fmt.Sprintf("%s%s", c.opts.endpoint, fmt.Sprintf(getBranchDetailPath, projectFullPath, branch, token))
	rsp, err := c.httpCli.DoGetRequest(uri, nil)
	if err != nil {
		return nil, fmt.Errorf("request git api error:%s", err.Error())
	}
	tGitBranch := &TGitBranch{}
	err = json.Unmarshal(rsp, tGitBranch)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal rsp failed:%s, rsp:%s", err.Error(), string(rsp))
	}
	return tGitBranch, nil
}

func (c *tGitClient) GetTagDetail(ctx context.Context, repoAddr, tag, token string) (*TGitTag, error) {
	projectFullPath := enCodeRepoFullpath(repoAddr)
	uri := fmt.Sprintf("%s%s", c.opts.endpoint, fmt.Sprintf(getTagPath, projectFullPath, tag, token))
	rsp, err := c.httpCli.DoGetRequest(uri, nil)
	if err != nil {
		return nil, fmt.Errorf("request git api error:%s", err.Error())
	}
	tGitTag := &TGitTag{}
	err = json.Unmarshal(rsp, tGitTag)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal rsp failed:%s, rsp:%s", err.Error(), string(rsp))
	}
	return tGitTag, nil
}
