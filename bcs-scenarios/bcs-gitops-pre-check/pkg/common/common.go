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

// Package common xxx
package common

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/argoproj/argo-cd/v2/util/db"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/pkg/apis/argo"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/pkg/apis/git"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/pkg/storage"
	precheck "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/proto"
)

// PublicFunc public func
type PublicFunc interface {
	GetMrInfo(ctx context.Context, repo, mrIID string) (*precheck.MRInfoData, error)
	RecordPreCheckTask(ctx context.Context, req *precheck.PreCheckTask) (*precheck.PreCheckTask, error)
	GetPreCheckTask(ctx context.Context, id int, project string) (*precheck.PreCheckTask, error)
	UpdatePreCheckTask(ctx context.Context, req *precheck.PreCheckTask) (*precheck.PreCheckTask, error)
	QueryPreCheckTaskList(ctx context.Context, query *precheck.ListTaskByIDReq) ([]*precheck.PreCheckTask, error)
}

type publicFunc struct {
	argoDBClient     argo.Client
	gitClientFactory git.Factory
	db               storage.Interface
	opts             *PublicFuncOpts
}

// PublicFuncOpts opts
type PublicFuncOpts struct {
	PowerAppEp string
}

// NewPublicFunc new func
func NewPublicFunc(argoDB db.ArgoDB, gitClientFactory git.Factory, db storage.Interface,
	opts *PublicFuncOpts) PublicFunc {
	return &publicFunc{
		argoDBClient:     argoDB,
		gitClientFactory: gitClientFactory,
		db:               db,
		opts:             opts,
	}
}

// GetMrInfo get mr info
func (p *publicFunc) GetMrInfo(ctx context.Context, repo, mrIID string) (*precheck.MRInfoData, error) {
	repoInfo, err := p.argoDBClient.GetRepository(ctx, repo)
	if err != nil {
		return nil, err
	}
	repoToken := repoInfo.Password
	gitClient, err := p.gitClientFactory.GetClient(repo)
	if err != nil {
		return nil, fmt.Errorf("get git client failed:%s", err.Error())
	}
	mrInfo, err := gitClient.GetMrInfo(ctx, repo, repoToken, mrIID)
	if err != nil {
		return nil, fmt.Errorf("get mr info failed:%s", err.Error())
	}
	return mrInfo, nil
}

// GetBranchInfo get branch info
func (p *publicFunc) GetBranchInfo(ctx context.Context, repo, branch string) (*git.TGitBranch, error) {
	repoInfo, err := p.argoDBClient.GetRepository(ctx, repo)
	if err != nil {
		return nil, err
	}
	repoToken := repoInfo.Password
	gitClient, err := p.gitClientFactory.GetClient(repo)
	if err != nil {
		return nil, fmt.Errorf("get git client failed:%s", err.Error())
	}
	branchInfo, err := gitClient.GetBranchDetail(ctx, repo, branch, repoToken)
	if err != nil {
		return nil, fmt.Errorf("get mr info failed:%s", err.Error())
	}
	return branchInfo, nil
}

// GetTagInfo get tag info
func (p *publicFunc) GetTagInfo(ctx context.Context, repo, tag string) (*git.TGitTag, error) {
	repoInfo, err := p.argoDBClient.GetRepository(ctx, repo)
	if err != nil {
		return nil, err
	}
	repoToken := repoInfo.Password
	gitClient, err := p.gitClientFactory.GetClient(repo)
	if err != nil {
		return nil, fmt.Errorf("get git client failed:%s", err.Error())
	}
	tagInfo, err := gitClient.GetTagDetail(ctx, repo, tag, repoToken)
	if err != nil {
		return nil, fmt.Errorf("get tag info failed:%s", err.Error())
	}
	return tagInfo, nil
}

// RecordPreCheckTask record
func (p *publicFunc) RecordPreCheckTask(ctx context.Context,
	req *precheck.PreCheckTask) (*precheck.PreCheckTask, error) {
	storageTask := storage.InitTask()
	storageTask.Project = req.Project
	storageTask.RepositoryAddr = req.RepositoryAddr
	storageTask.MrIID = req.MrIid
	storageTask.CheckCallbackGit = *req.CheckCallbackGit
	storageTask.CheckRevision = req.CheckRevision
	storageTask.ApplicationName = req.ApplicationName
	storageTask.TriggerType = req.TriggerType
	storageTask.BranchValue = req.BranchValue
	storageTask.CreateBy = "plugin"
	storageTask.TriggerByUser = req.TriggerByUser
	storageTask.FlowID = req.FlowID
	storageTask.FlowLink = req.FlowLink
	storageTask.NeedReplaceRepo = *req.NeedReplaceRepo
	storageTask.ReplaceRepo = req.ReplaceRepo
	storageTask.ReplaceProject = req.ReplaceProject
	storageTask.ChooseApplication = *req.ChooseApplication
	storageTask.AppFilter = req.AppFilter
	storageTask.LabelSelector = req.LabelSelector
	p.getRevision(ctx, storageTask, req)
	task, err := p.db.CreatePreCheckTask(storageTask)
	if err != nil {
		return nil, err
	}
	protoTask, err := p.transStorageTaskToProto(task)
	if err != nil {
		blog.Errorf("trans task failed:%s", err.Error())
		return nil, err
	}
	if *req.CheckCallbackGit {
		go p.LockMR(ctx, protoTask)
	}
	return protoTask, nil
}

// GetPreCheckTask get task
func (p *publicFunc) GetPreCheckTask(ctx context.Context, id int, project string) (*precheck.PreCheckTask, error) {
	storageTask, err := p.db.GetPreCheckTask(id, project)
	if err != nil {
		return nil, err
	}
	if storageTask == nil {
		return nil, fmt.Errorf("can not find task, id:%d, project:%s", id, project)
	}
	return p.transStorageTaskToProto(storageTask)
}

// UpdatePreCheckTask update task
func (p *publicFunc) UpdatePreCheckTask(ctx context.Context,
	req *precheck.PreCheckTask) (*precheck.PreCheckTask, error) {
	id, err := strconv.Atoi(req.Id)
	if err != nil {
		return nil, fmt.Errorf("id illeagal:%s", req.Id)
	}
	storageTask, err := p.db.GetPreCheckTask(id, "")
	if err != nil {
		return nil, fmt.Errorf("get task by id %d failed:%s", id, err.Error())
	}
	if storageTask == nil {
		return nil, fmt.Errorf("cannot find task by id %d", id)
	}
	mergeTaskToStorage(req, storageTask)
	storageTask.UpdateTime = time.Now().UTC()
	if storageTask.CheckCallbackGit && storageTask.Finish && storageTask.MrIID != "" {
		go p.CommentMR(ctx, req)
		go p.UpdateMRCheckStatus(ctx, req)
	}
	err = p.db.UpdatePreCheckTask(storageTask)
	if err != nil {
		return nil, fmt.Errorf("update task failed:%s", err.Error())
	}
	return p.GetPreCheckTask(ctx, id, storageTask.Project)
}

// nolint
func mergeTaskToStorage(protoTask *precheck.PreCheckTask, storageTask *storage.PreCheckTask) {
	handleStringType(protoTask, storageTask)
	handleJsonStr(protoTask, storageTask)
	handleBoolType(protoTask, storageTask)
}

// nolint
func handleStringType(protoTask *precheck.PreCheckTask, storageTask *storage.PreCheckTask) {
	if storageTask.Project != protoTask.Project && protoTask.Project != "" {
		storageTask.Project = protoTask.Project
	}
	if storageTask.RepositoryAddr != protoTask.RepositoryAddr && protoTask.RepositoryAddr != "" {
		storageTask.RepositoryAddr = protoTask.RepositoryAddr
	}
	if storageTask.MrIID != protoTask.MrIid && protoTask.MrIid != "" {
		storageTask.MrIID = protoTask.MrIid
	}
	if storageTask.CheckRevision != protoTask.CheckRevision && protoTask.CheckRevision != "" {
		storageTask.CheckRevision = protoTask.CheckRevision
	}
	if storageTask.ApplicationName != protoTask.ApplicationName && protoTask.ApplicationName != "" {
		storageTask.ApplicationName = protoTask.ApplicationName
	}
	if storageTask.TriggerType != protoTask.TriggerType && protoTask.TriggerType != "" {
		storageTask.TriggerType = protoTask.TriggerType
	}
	if storageTask.BranchValue != protoTask.BranchValue && protoTask.BranchValue != "" {
		storageTask.BranchValue = protoTask.BranchValue
	}
	if storageTask.TriggerByUser != protoTask.TriggerByUser && protoTask.TriggerByUser != "" {
		storageTask.TriggerByUser = protoTask.TriggerByUser
	}
	if storageTask.CreateBy != protoTask.CreateBy && protoTask.CreateBy != "" {
		storageTask.CreateBy = protoTask.CreateBy
	}
	if storageTask.FlowID != protoTask.FlowID && protoTask.FlowID != "" {
		storageTask.FlowID = protoTask.FlowID
	}
	if storageTask.ReplaceRepo != protoTask.ReplaceRepo && protoTask.ReplaceRepo != "" {
		storageTask.ReplaceRepo = protoTask.ReplaceRepo
	}
	if storageTask.ReplaceProject != protoTask.ReplaceProject && protoTask.ReplaceProject != "" {
		storageTask.ReplaceProject = protoTask.ReplaceProject
	}
	if storageTask.FlowLink != protoTask.FlowLink && protoTask.FlowLink != "" {
		storageTask.FlowLink = protoTask.FlowLink
	}
	if storageTask.Message != protoTask.Message && protoTask.Message != "" {
		storageTask.Message = protoTask.Message
	}
	if storageTask.AppFilter != protoTask.AppFilter && protoTask.AppFilter != "" {
		storageTask.AppFilter = protoTask.AppFilter
	}
	if storageTask.LabelSelector != protoTask.LabelSelector && protoTask.LabelSelector != "" {
		storageTask.LabelSelector = protoTask.LabelSelector
	}
}

func handleJsonStr(protoTask *precheck.PreCheckTask, storageTask *storage.PreCheckTask) {
	if len(protoTask.InvolvedApplications) != 0 {
		involvedApplicationsStr, _ := json.Marshal(protoTask.InvolvedApplications)
		storageTask.InvolvedApplications = string(involvedApplicationsStr)
	}
	if len(protoTask.CheckDetail) != 0 {
		checkDetailStr, _ := json.Marshal(protoTask.CheckDetail)
		storageTask.CheckDetail = string(checkDetailStr)
	}
	if protoTask.MrInfo != nil {
		mrInfoStr, _ := json.Marshal(protoTask.MrInfo)
		storageTask.MrInfo = string(mrInfoStr)
	}
}

func handleBoolType(protoTask *precheck.PreCheckTask, storageTask *storage.PreCheckTask) {
	if protoTask.CheckCallbackGit != nil {
		storageTask.CheckCallbackGit = *protoTask.CheckCallbackGit
	}
	if protoTask.Finish != nil {
		storageTask.Finish = *protoTask.Finish
	}
	if protoTask.NeedReplaceRepo != nil {
		storageTask.NeedReplaceRepo = *protoTask.NeedReplaceRepo
	}
	if protoTask.Pass != nil {
		storageTask.Pass = *protoTask.Pass
	}
	if protoTask.ChooseApplication != nil {
		storageTask.ChooseApplication = *protoTask.ChooseApplication
	}
}

// LockMR lock mr
func (p *publicFunc) LockMR(ctx context.Context, task *precheck.PreCheckTask) {
	repoInfo, err := p.argoDBClient.GetRepository(ctx, task.RepositoryAddr)
	if err != nil {
		blog.Errorf("get token failed:%s", err.Error())
		return
	}
	repoToken := repoInfo.Password
	gitClient, err := p.gitClientFactory.GetClient(task.RepositoryAddr)
	if err != nil {
		blog.Errorf("get git client failed:%s", err.Error())
		return
	}
	err = gitClient.SubmitCheckState(ctx, "powerapp", "pending", task.FlowLink, "gitops部署前检查",
		repoToken, true, task)
	if err != nil {
		blog.Errorf("comment failed:%s", err.Error())
		return
	}
}

// UpdateMRCheckStatus update status
func (p *publicFunc) UpdateMRCheckStatus(ctx context.Context, task *precheck.PreCheckTask) {
	repoInfo, err := p.argoDBClient.GetRepository(ctx, task.RepositoryAddr)
	if err != nil {
		blog.Errorf("get token failed:%s", err.Error())
		return
	}
	repoToken := repoInfo.Password
	gitClient, err := p.gitClientFactory.GetClient(task.RepositoryAddr)
	if err != nil {
		blog.Errorf("get git client failed:%s", err.Error())
		return
	}
	state := "success"
	block := false
	if !*task.Pass {
		state = "failure"
		block = true
	}
	err = gitClient.SubmitCheckState(ctx, "powerapp", state, task.FlowLink, "gitops部署前检查",
		repoToken, block, task)
	if err != nil {
		blog.Errorf("comment failed:%s", err.Error())
		return
	}
}

// CommentMR comment mr
func (p *publicFunc) CommentMR(ctx context.Context, task *precheck.PreCheckTask) {
	repoInfo, err := p.argoDBClient.GetRepository(ctx, task.RepositoryAddr)
	if err != nil {
		blog.Errorf("get token failed:%s", err.Error())
		return
	}
	repoToken := repoInfo.Password
	gitClient, err := p.gitClientFactory.GetClient(task.RepositoryAddr)
	if err != nil {
		blog.Errorf("get git client failed:%s", err.Error())
		return
	}
	linkComment := fmt.Sprintf("检查详情：%s", p.opts.PowerAppEp)
	err = gitClient.CommentMR(ctx, task.RepositoryAddr, repoToken, task.MrIid, linkComment)
	if err != nil {
		blog.Errorf("comment failed:%s", err.Error())
		return
	}
	// diffFinish := true
	// diffPass := true
	// for step := range task.CheckDetail {
	//	switch step {
	//	case "diff":
	//		for app := range task.CheckDetail[step].CheckDetail {
	//			if !task.CheckDetail[step].CheckDetail[app].Finish {
	//				diffFinish = false
	//			}
	//		}
	//	}
	// }
}

// QueryPreCheckTaskList query tasks
func (p *publicFunc) QueryPreCheckTaskList(ctx context.Context,
	query *precheck.ListTaskByIDReq) ([]*precheck.PreCheckTask, error) {
	storageQuery := &storage.PreCheckTaskQuery{
		Projects:     query.Projects,
		Repositories: query.Repos,
		StartTime:    query.StartTime,
		EndTime:      query.EndTime,
		Limit:        int(query.Limit),
		Offset:       int(query.Offset),
		WithDetail:   query.WithDetail,
	}
	if storageQuery.Limit == 0 {
		storageQuery.Limit = 50
	}
	result, err := p.db.ListPreCheckTask(storageQuery)
	if err != nil {
		return nil, fmt.Errorf("query task list failed:%s", err.Error())
	}
	blog.Infof("query:%v", storageQuery)
	blog.Infof("result:%v", result)
	projectMap := make(map[string]bool)
	for _, project := range query.Projects {
		projectMap[project] = true
	}
	protoTaskList := make([]*precheck.PreCheckTask, 0)
	for _, task := range result {
		if task.NeedReplaceRepo && task.ReplaceProject != "" && !projectMap[task.ReplaceProject] {
			blog.Infof("task %d is inplace by project %s, skip.", task.ID, task.ReplaceProject)
			continue
		}
		protoTask, err := p.transStorageTaskToProto(task)
		if err != nil {
			blog.Errorf("trans task failed:%s", err.Error())
			return nil, err
		}
		protoTaskList = append(protoTaskList, protoTask)
	}
	return protoTaskList, nil
}

func (p *publicFunc) transStorageTaskToProto(storageTask *storage.PreCheckTask) (*precheck.PreCheckTask, error) {
	protoTask := &precheck.PreCheckTask{
		Id:                strconv.Itoa(storageTask.ID),
		Project:           storageTask.Project,
		RepositoryAddr:    storageTask.RepositoryAddr,
		MrIid:             storageTask.MrIID,
		CheckCallbackGit:  &storageTask.CheckCallbackGit,
		CheckRevision:     storageTask.CheckRevision,
		ApplicationName:   storageTask.ApplicationName,
		TriggerType:       storageTask.TriggerType,
		BranchValue:       storageTask.BranchValue,
		CreateTime:        storageTask.CreateTime.UTC().Format("2006-01-02T15:04:05Z"),
		UpdateTime:        storageTask.UpdateTime.UTC().Format("2006-01-02T15:04:05Z"),
		TriggerByUser:     storageTask.TriggerByUser,
		CreateBy:          storageTask.CreateBy,
		Finish:            &storageTask.Finish,
		FlowID:            storageTask.FlowID,
		ReplaceRepo:       storageTask.ReplaceRepo,
		NeedReplaceRepo:   &storageTask.NeedReplaceRepo,
		ReplaceProject:    storageTask.ReplaceProject,
		FlowLink:          storageTask.FlowLink,
		Pass:              &storageTask.Pass,
		Message:           storageTask.Message,
		ChooseApplication: &storageTask.ChooseApplication,
		AppFilter:         storageTask.AppFilter,
		LabelSelector:     storageTask.LabelSelector,
	}

	checkDetail := make(map[string]*precheck.ApplicationCheckDetail)
	if storageTask.CheckDetail != "" {
		if unmarshalErr := json.Unmarshal([]byte(storageTask.CheckDetail), &checkDetail); unmarshalErr != nil {
			return nil, fmt.Errorf("unmashal checkDetail fail:%s", unmarshalErr.Error())
		}
	}
	protoTask.CheckDetail = checkDetail
	involvedApplications := make([]string, 0)
	if storageTask.InvolvedApplications != "" {
		if unmarshalErr := json.Unmarshal([]byte(storageTask.InvolvedApplications),
			&involvedApplications); unmarshalErr != nil {
			return nil, fmt.Errorf("unmashal involvedApplications fail:%s", unmarshalErr.Error())
		}
	}
	protoTask.InvolvedApplications = involvedApplications
	mrInfo := &precheck.MRInfoData{}
	if storageTask.MrInfo != "" {
		if unmarshalErr := json.Unmarshal([]byte(storageTask.MrInfo), &mrInfo); unmarshalErr != nil {
			return nil, fmt.Errorf("unmashal mrInfo fail:%s", unmarshalErr.Error())
		}
	}
	protoTask.MrInfo = mrInfo
	return protoTask, nil
}

func (p *publicFunc) getRevision(ctx context.Context, storageTask *storage.PreCheckTask, req *precheck.PreCheckTask) {
	switch req.TriggerType {
	case "mr":
		mrInfo, err := p.GetMrInfo(ctx, req.RepositoryAddr, req.MrIid)
		if err != nil {
			blog.Errorf("get mr info err:%s", err.Error())
			req.CheckRevision = ""
			storageTask.CheckRevision = ""
			storageTask.Message = fmt.Sprintf("get mr info err:%s", err.Error())
		} else {
			req.CheckRevision = mrInfo.SourceCommit
			req.MrInfo = mrInfo
			storageTask.CheckRevision = mrInfo.SourceCommit
			mrInfoStr, _ := json.Marshal(mrInfo)
			storageTask.MrInfo = string(mrInfoStr)
		}
		return
	case "commit":
		req.CheckRevision = req.BranchValue
		storageTask.CheckRevision = req.BranchValue
	case "tag":
		tagInfo, err := p.GetTagInfo(ctx, req.RepositoryAddr, req.BranchValue)
		if err != nil {
			blog.Errorf("get tag info err:%s", err.Error())
			req.CheckRevision = ""
			storageTask.CheckRevision = ""
			storageTask.Message = fmt.Sprintf("get tag info err:%s", err.Error())
		} else {
			req.CheckRevision = tagInfo.Commit.ID
			storageTask.CheckRevision = tagInfo.Commit.ID
		}
	case "branch":
		branchInfo, err := p.GetBranchInfo(ctx, req.RepositoryAddr, req.BranchValue)
		if err != nil {
			blog.Errorf("get branch info err:%s", err.Error())
			req.CheckRevision = ""
			storageTask.CheckRevision = ""
			storageTask.Message = fmt.Sprintf("get branch info err:%s", err.Error())
		} else {
			req.CheckRevision = branchInfo.Commit.ID
			storageTask.CheckRevision = branchInfo.Commit.ID
		}
	}
	blog.Info(storageTask.CheckRevision)
}
