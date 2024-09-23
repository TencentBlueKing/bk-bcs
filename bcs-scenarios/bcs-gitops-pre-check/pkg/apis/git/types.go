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
	"net/url"
	"strings"
)

// TGitBranch branch
type TGitBranch struct {
	Protected          bool       `json:"protected"`
	DevelopersCanPush  bool       `json:"developers_can_push"`
	DevelopersCanMerge bool       `json:"developers_can_merge"`
	Name               string     `json:"name"`
	Commit             TGitCommit `json:"commit"`
	Description        string     `json:"description"`
	CreatedAt          string     `json:"created_at"`
	BranchType         BranchType `json:"branch_type"`
	Author             Author     `json:"author"`
}

// TGitCommit commit
type TGitCommit struct {
	ID             string      `json:"id"`
	Message        string      `json:"message"`
	ParentIDs      []string    `json:"parent_ids"`
	AuthoredDate   string      `json:"authored_date"`
	AuthorName     string      `json:"author_name"`
	AuthorEmail    string      `json:"author_email"`
	CommittedDate  string      `json:"committed_date"`
	CommitterName  string      `json:"committer_name"`
	CommitterEmail string      `json:"committer_email"`
	Title          string      `json:"title"`
	ScrollObjectID interface{} `json:"scroll_object_id"`
	CreatedAt      string      `json:"created_at"`
	ShortID        string      `json:"short_id"`
}

// TGitTag tag
type TGitTag struct {
	Name        string     `json:"name"`
	Message     string     `json:"message"`
	Commit      TGitCommit `json:"commit"`
	CreatedAt   string     `json:"created_at"`
	Description string     `json:"description"`
}

// BranchType type
type BranchType struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Author author
type Author struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	WebURL    string `json:"web_url"`
	Name      string `json:"name"`
	State     string `json:"state"`
	AvatarURL string `json:"avatar_url"`
}

type tGitMrInfoResp struct {
	ID         int    `json:"id"`
	IID        int    `json:"iid"`
	Title      string `json:"title"`
	CreateTime string `json:"created_at"`
	UpdateTime string `json:"updated_at"`
	Author     struct {
		Username string `json:"username"`
	} `json:"author"`
	Description  string `json:"description"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Repository   string `json:"repository"`
	SourceCommit string `json:"source_commit"`
	TargetCommit string `json:"target_commit"`
}

type tGitMrCommentResp struct {
	CreateAt string `json:"created_at"`
	UpdateAt string `json:"updated_at"`
	Author   struct {
		Username string `json:"username"`
	} `json:"author"`
	Reviewers        []interface{} `json:"reviewers"`
	ID               int           `json:"id"`
	ProjectID        int           `json:"project_id"`
	ReviewableID     int           `json:"reviewable_id"`
	State            string        `json:"state"`
	RestrictType     string        `json:"restrict_type"`
	PushResetEnabled bool          `json:"push_reset_enabled"`
}

// enCodeRepoFullpath repourl -> repofullpath
func enCodeRepoFullpath(repoUrl string) string {
	parsedUrl, err := url.Parse(repoUrl)
	if err != nil {
		return repoUrl
	}
	return strings.Trim(url.QueryEscape(strings.TrimSuffix(parsedUrl.RequestURI(), ".git")), "%2F")
}
