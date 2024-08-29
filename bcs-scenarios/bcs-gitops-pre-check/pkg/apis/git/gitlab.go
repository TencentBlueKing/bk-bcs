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
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/pkg/apis/requester"
	precheck "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/proto"
)

type gitLabClient struct {
	httpCli requester.Requester
	opts    *clientOpts
}

var (
	// ErrNotSupportCheckState not support
	ErrNotSupportCheckState = errors.New("current git client is not support submit checkstate")
)

// NewGitLab new gitlab client
func NewGitLab(opts *clientOpts) Client {

	return &gitLabClient{opts: opts, httpCli: requester.NewRequester()}
}

// GetMrInfo get mr info
func (c *gitLabClient) GetMrInfo(ctx context.Context, repoAddr, token, mrIID string) (*precheck.MRInfoData, error) {
	cli, err := gitlab.NewClient(token, gitlab.WithBaseURL(c.opts.endpoint))
	if err != nil {
		return nil, fmt.Errorf("new gitlab client failed:%s", err.Error())
	}
	projectFullPath := enCodeRepoFullpath(repoAddr)
	intMrIid, err := strconv.Atoi(mrIID)
	if err != nil {
		return nil, fmt.Errorf("trans mrIId %s to int failed:%s", mrIID, err.Error())
	}
	mr, _, err := cli.MergeRequests.GetMergeRequest(projectFullPath, intMrIid, &gitlab.GetMergeRequestsOptions{})
	if err != nil {
		blog.Infof("get gitlab mr request failed:%s", err.Error())
		return nil, fmt.Errorf("get gitlab mr request failed:%s", err.Error())
	}
	mrInfo := &precheck.MRInfoData{
		SourceBranch: mr.SourceBranch,
		TargetBranch: mr.TargetBranch,
		Creator:      mr.Author.Name,
		CreateTime:   mr.CreatedAt.String(),
		UpdateTime:   mr.UpdatedAt.String(),
		Title:        mr.Title,
		MrMessage:    mr.Description,
		Repository:   repoAddr,
		SourceCommit: mr.SourceBranch,
		TargetCommit: mr.TargetBranch,
	}
	return mrInfo, nil
}

// CommentMR comment
func (c *gitLabClient) CommentMR(ctx context.Context, repoAddr, token, mrIID, comment string) error {
	cli, err := gitlab.NewClient(token, gitlab.WithBaseURL(c.opts.endpoint))
	if err != nil {
		return fmt.Errorf("new gitlab client failed:%s", err.Error())
	}
	projectFullPath := enCodeRepoFullpath(repoAddr)
	intMrIid, _ := strconv.Atoi(mrIID)
	_, _, err = cli.Notes.CreateMergeRequestNote(projectFullPath, intMrIid, &gitlab.CreateMergeRequestNoteOptions{
		Body: &comment,
	})
	if err != nil {
		return err
	}
	return nil
}

// SubmitCheckState submit state
func (c *gitLabClient) SubmitCheckState(ctx context.Context, checkSystem, state, targetUrl, description,
	token string, block bool, req *precheck.PreCheckTask) error {
	return ErrNotSupportCheckState
}

func (c *gitLabClient) GetBranchDetail(ctx context.Context, repoAddr, branch, token string) (*TGitBranch, error) {
	cli, err := gitlab.NewClient(token, gitlab.WithBaseURL(c.opts.endpoint))
	if err != nil {
		return nil, fmt.Errorf("new gitlab client failed:%s", err.Error())
	}
	projectFullPath := enCodeRepoFullpath(repoAddr)
	branchInfo, _, err := cli.Branches.GetBranch(projectFullPath, branch)
	if err != nil {
		blog.Infof("get gitlab branch request failed:%s", err.Error())
		return nil, fmt.Errorf("get gitlab branch request failed:%s", err.Error())
	}
	resBranchInfo := &TGitBranch{
		Protected:          branchInfo.Protected,
		DevelopersCanPush:  branchInfo.DevelopersCanPush,
		DevelopersCanMerge: branchInfo.DevelopersCanMerge,
		Name:               branchInfo.Name,
		Commit: TGitCommit{
			ID:             branchInfo.Commit.ID,
			Message:        branchInfo.Commit.Message,
			ParentIDs:      branchInfo.Commit.ParentIDs,
			AuthoredDate:   branchInfo.Commit.AuthoredDate.String(),
			AuthorName:     branchInfo.Commit.AuthorName,
			AuthorEmail:    branchInfo.Commit.AuthorEmail,
			CommittedDate:  branchInfo.Commit.CommittedDate.String(),
			CommitterName:  branchInfo.Commit.CommitterName,
			CommitterEmail: branchInfo.Commit.CommitterEmail,
			Title:          branchInfo.Commit.Title,
			ScrollObjectID: "",
			CreatedAt:      branchInfo.Commit.CreatedAt.String(),
			ShortID:        branchInfo.Commit.ShortID,
		},
	}
	return resBranchInfo, nil
}

func (c *gitLabClient) GetTagDetail(ctx context.Context, repoAddr, tag, token string) (*TGitTag, error) {
	cli, err := gitlab.NewClient(token, gitlab.WithBaseURL(c.opts.endpoint))
	if err != nil {
		return nil, fmt.Errorf("new gitlab client failed:%s", err.Error())
	}
	projectFullPath := enCodeRepoFullpath(repoAddr)
	tagInfo, _, err := cli.Tags.GetTag(projectFullPath, tag)
	if err != nil {
		blog.Infof("get gitlab tag request failed:%s", err.Error())
		return nil, fmt.Errorf("get gitlab tag request failed:%s", err.Error())
	}
	resTagInfo := &TGitTag{
		Name:    tagInfo.Name,
		Message: tagInfo.Message,
		Commit: TGitCommit{
			ID:             tagInfo.Commit.ID,
			Message:        tagInfo.Commit.Message,
			ParentIDs:      tagInfo.Commit.ParentIDs,
			AuthoredDate:   tagInfo.Commit.AuthoredDate.String(),
			AuthorName:     tagInfo.Commit.AuthorName,
			AuthorEmail:    tagInfo.Commit.AuthorEmail,
			CommittedDate:  tagInfo.Commit.CommittedDate.String(),
			CommitterName:  tagInfo.Commit.CommitterName,
			CommitterEmail: tagInfo.Commit.CommitterEmail,
			Title:          tagInfo.Commit.Title,
			ScrollObjectID: "",
			CreatedAt:      tagInfo.Commit.CreatedAt.String(),
			ShortID:        tagInfo.Commit.ShortID,
		},
		CreatedAt:   "",
		Description: "",
	}

	return resTagInfo, nil
}
