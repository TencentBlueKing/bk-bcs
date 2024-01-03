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

// Package auth NOTES
package auth

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	bkiam "github.com/TencentBlueKing/iam-go-sdk"
	"github.com/pkg/errors"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/auth-server/options"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sdk/auth"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sys"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbas "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/auth-server"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/space"
)

// Auth related operate.
type Auth struct {
	// auth related operate.
	auth auth.Authorizer
	// ds data service's auth related api.
	ds pbds.DataClient
	// disableAuth defines whether iam authorization is disabled
	disableAuth bool
	// disableWriteOpt defines which biz's write operation needs to be disabled
	disableWriteOpt *options.DisableWriteOption
	// iamSettings defines iam settings
	iamClient *bkiam.IAM
	// spaceMgr defines space manager
	spaceMgr *space.Manager
}

// NewAuth new auth.
func NewAuth(auth auth.Authorizer, ds pbds.DataClient, disableAuth bool, iamClient *bkiam.IAM,
	disableWriteOpt *options.DisableWriteOption, spaceMgr *space.Manager) (*Auth, error) {

	if auth == nil {
		return nil, errf.New(errf.InvalidParameter, "auth is nil")
	}

	if ds == nil {
		return nil, errf.New(errf.InvalidParameter, "data client is nil")
	}

	if disableWriteOpt == nil {
		return nil, errf.New(errf.InvalidParameter, "disable write operation is nil")
	}

	i := &Auth{
		auth:            auth,
		ds:              ds,
		disableAuth:     disableAuth,
		iamClient:       iamClient,
		disableWriteOpt: disableWriteOpt,
		spaceMgr:        spaceMgr,
	}

	return i, nil
}

// AuthorizeBatch authorize resource batch.
func (a *Auth) AuthorizeBatch(ctx context.Context, req *pbas.AuthorizeBatchReq) (*pbas.AuthorizeBatchResp, error) {
	kt := kit.FromGrpcContext(ctx)
	resp := new(pbas.AuthorizeBatchResp)

	if len(req.Resources) == 0 {
		resp.Decisions = make([]*pbas.Decision, 0)
		return resp, nil
	}

	// if write operations are disabled, returns corresponding error
	if err := a.isWriteOperationDisabled(kt, req.Resources); err != nil {
		return nil, err
	}

	// if auth is disabled, returns authorized for all request resources
	// if a.disableAuth {
	// 	resp.Decisions = make([]*pbas.Decision, len(req.Resources))
	// 	for index := range req.Resources {
	// 		resp.Decisions[index] = &pbas.Decision{Authorized: true}
	// 	}
	// 	return resp, nil
	// }

	// parse bscp resource to iam resource
	resources := pbas.ResourceAttributes(req.Resources)
	opts, decisions, err := parseAttributesToBatchOptions(kt, req.User.UserInfo(), resources...)
	if err != nil {
		return nil, err
	}

	// all resources are skipped
	if opts == nil {
		resp.Decisions = pbas.PbDecisions(decisions)
		return resp, nil
	}

	// do authentication
	authDecisions, err := a.auth.AuthorizeBatch(ctx, opts)
	if err != nil {
		logs.Errorf("authorize batch failed, ops: %#v, req: %#v, err: %v, rid: %s", err, opts, req, kt.Rid)
		return nil, err
	}

	index := 0
	decisionLen := len(decisions)
	for _, decision := range authDecisions {
		// skip resources' decisions are already set as authorized
		for index < decisionLen && decisions[index].Authorized {
			index++
		}

		if index >= decisionLen {
			break
		}

		decisions[index].Authorized = decision.Authorized
		index++
	}

	resp.Decisions = pbas.PbDecisions(decisions)
	return resp, nil
}

func (a *Auth) isWriteOperationDisabled(kt *kit.Kit, resources []*pbas.ResourceAttribute) error {
	if !a.disableWriteOpt.IsDisabled {
		return nil
	}

	for _, resource := range resources {
		action := meta.Action(resource.Basic.Action)
		if action == meta.Find || action == meta.SkipAction {
			continue
		}

		if a.disableWriteOpt.IsAll {
			logs.Errorf("all %s operation is disabled, rid: %s", action, kt.Rid)
			return errf.New(errf.Aborted, "bscp server is publishing, wring operation is not allowed")
		}

		bizID := resource.BizId
		if _, exists := a.disableWriteOpt.BizIDMap.Load(bizID); exists {
			logs.Errorf("biz id %d %s operation is disabled, rid: %s", bizID, action, kt.Rid)
			return errf.New(errf.Aborted, "bscp server is publishing, wring operation is not allowed")
		}
	}

	return nil
}

// parseAttributesToBatchOptions parse auth attributes to authorize batch options
func parseAttributesToBatchOptions(kt *kit.Kit, user *meta.UserInfo, resources ...*meta.ResourceAttribute) (
	*client.AuthBatchOptions, []*meta.Decision, error) {

	authBatchArr := make([]*client.AuthBatch, 0)
	decisions := make([]*meta.Decision, len(resources))
	for index, resource := range resources {
		decisions[index] = &meta.Decision{
			Resource:   resource,
			Authorized: false,
		}

		// this resource should be skipped, do not need to verify in auth center.
		if resource.Basic.Action == meta.SkipAction {
			decisions[index].Authorized = true
			logs.V(5).Infof("skip authorization for resource: %+v, rid: %s", resource, kt.Rid)
			continue
		}

		action, iamResources, err := AdaptAuthOptions(resource)
		if err != nil {
			logs.Errorf("adapt bscp resource to iam failed, err: %s, rid: %s", err, kt.Rid)
			return nil, nil, err
		}

		// this resource should be skipped, do not need to verify in auth center.
		if action == sys.Skip {
			decisions[index].Authorized = true
			logs.V(5).Infof("skip authorization for resource: %+v, rid: %s", resource, kt.Rid)
			continue
		}

		authBatchArr = append(authBatchArr, &client.AuthBatch{
			Action:    client.Action{ID: string(action)},
			Resources: iamResources,
		})
	}

	// all resources are skipped
	if len(authBatchArr) == 0 {
		return nil, decisions, nil
	}

	ops := &client.AuthBatchOptions{
		System: sys.SystemIDBSCP,
		Subject: client.Subject{
			Type: "user",
			ID:   user.UserName,
		},
		Batch: authBatchArr,
	}
	return ops, decisions, nil
}

// GetPermissionToApply get iam permission to apply when user has no permission to some resources.
func (a *Auth) GetPermissionToApply(ctx context.Context, req *pbas.GetPermissionToApplyReq) (
	*pbas.GetPermissionToApplyResp, error) {

	kt := kit.FromGrpcContext(ctx)
	resp := new(pbas.GetPermissionToApplyResp)

	permission, err := a.getPermissionToApply(kt, pbas.ResourceAttributes(req.Resources))
	if err != nil {
		return nil, err
	}

	resourceAttributes := pbas.ResourceAttributes(req.Resources)
	application, err := AdaptIAMApplicationOptions(resourceAttributes)
	if err != nil {
		return nil, err
	}
	url, err := a.iamClient.GetApplyURL(*application, "", kt.User)
	if err != nil {
		return nil, errors.Wrap(err, "gen apply url")
	}
	resp.ApplyUrl = url

	resp.Permission = pbas.PbIamPermission(permission)
	return resp, nil
}

func (a *Auth) getPermissionToApply(kt *kit.Kit, resources []*meta.ResourceAttribute) (*meta.IamPermission, error) {
	permission := new(meta.IamPermission)
	permission.SystemID = sys.SystemIDBSCP
	permission.SystemName = sys.SystemNameBSCP

	// parse bscp auth resource
	resTypeIDsMap, permissionMap, err := a.parseResources(kt, resources)
	if err != nil {
		logs.Errorf("get inst ID and name map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get bscp resource name by id, then assign it to corresponding iam auth resource
	instIDNameMap, err := a.getInstIDNameMap(kt, resTypeIDsMap)
	if err != nil {
		return nil, err
	}

	for actionID, permissionTypeMap := range permissionMap {
		action := &meta.IamAction{
			ID:                   string(actionID),
			Name:                 sys.ActionIDNameMap[actionID],
			RelatedResourceTypes: make([]*meta.IamResourceType, 0),
		}

		for rscType := range permissionTypeMap {
			iamResourceType := permissionTypeMap[rscType]

			for idx, resources := range iamResourceType.Instances {
				for idx2, resource := range resources {
					iamResourceType.Instances[idx][idx2].Name = instIDNameMap[resource.ID]
				}
			}

			action.RelatedResourceTypes = append(action.RelatedResourceTypes, iamResourceType)
		}
		permission.Actions = append(permission.Actions, action)
	}

	return permission, nil
}

// parseResources parse bscp auth resource to iam permission resources in organized way
func (a *Auth) parseResources(kt *kit.Kit, resources []*meta.ResourceAttribute) (map[client.TypeID][]string,
	map[client.ActionID]map[client.TypeID]*meta.IamResourceType, error) {

	// resTypeIDsMap maps resource type to resource ids to get resource names.
	resTypeIDsMap := make(map[client.TypeID][]string)
	// permissionMap maps ActionID and TypeID to ResourceInstances
	permissionMap := make(map[client.ActionID]map[client.TypeID]*meta.IamResourceType, 0)

	for _, r := range resources {
		// parse bscp auth resource to iam action id and iam resources
		actionID, resources, err := AdaptAuthOptions(r)
		if err != nil {
			logs.Errorf("adaptor bscp resource to iam failed, err: %s, rid: %s", err, kt.Rid)
			return nil, nil, err
		}

		if _, ok := permissionMap[actionID]; !ok {
			permissionMap[actionID] = make(map[client.TypeID]*meta.IamResourceType, 0)
		}

		// generate iam resource resources by its paths and itself
		for _, res := range resources {
			if len(res.ID) == 0 && res.Attribute == nil {
				continue
			}

			resTypeIDsMap[res.Type] = append(resTypeIDsMap[res.Type], res.ID)

			resource := make([]*meta.IamResourceInstance, 0)
			if res.Attribute != nil {
				// parse bscp auth resource iam path attribute to iam ancestor resources
				iamPath, ok := res.Attribute[client.IamPathKey].([]string)
				if !ok {
					return nil, nil, fmt.Errorf("iam path(%v) is not string array", res.Attribute[client.IamPathKey])
				}

				ancestors, err := a.parseIamPathToAncestors(iamPath)
				if err != nil {
					return nil, nil, err
				}
				resource = append(resource, ancestors...)

				// record ancestor resource ids to get names from them afterwards
				for _, ancestor := range ancestors {
					ancestorType := client.TypeID(ancestor.Type)
					resTypeIDsMap[ancestorType] = append(resTypeIDsMap[ancestorType], ancestor.ID)
				}
			}

			// add iam resource of auth resource to the related iam resources after its ancestors
			resource = append(resource, &meta.IamResourceInstance{
				Type:     string(res.Type),
				TypeName: sys.ResourceTypeIDMap[res.Type],
				ID:       res.ID,
			})

			if permissionMap[actionID][res.Type] == nil {
				permissionMap[actionID][res.Type] = &meta.IamResourceType{
					SystemID:   res.System,
					SystemName: sys.SystemIDNameMap[res.System],
					Type:       string(res.Type),
					TypeName:   sys.ResourceTypeIDMap[res.Type],
					Instances:  make([][]*meta.IamResourceInstance, 0),
				}
			}
			permissionMap[actionID][res.Type].Instances = append(permissionMap[actionID][res.Type].Instances, resource)
		}
	}

	return resTypeIDsMap, permissionMap, nil
}

// parseIamPathToAncestors parse iam path to resource's ancestor resources
func (a *Auth) parseIamPathToAncestors(iamPath []string) ([]*meta.IamResourceInstance, error) {
	resources := make([]*meta.IamResourceInstance, 0)
	for _, path := range iamPath {
		pathItemArr := strings.Split(strings.Trim(path, "/"), "/")
		for _, pathItem := range pathItemArr {
			typeAndID := strings.Split(pathItem, ",")
			if len(typeAndID) != 2 {
				return nil, fmt.Errorf("pathItem %s invalid", pathItem)
			}
			id := typeAndID[1]
			if id == "*" {
				continue
			}
			resources = append(resources, &meta.IamResourceInstance{
				Type:     typeAndID[0],
				TypeName: sys.ResourceTypeIDMap[client.TypeID(typeAndID[0])],
				ID:       id,
			})
		}
	}
	return resources, nil
}

// Note how to get ancestor names? right now it means cc biz name, which is not in bscp
// getInstIDNameMap get resource id to name map by resource ids, groups by resource type
func (a *Auth) getInstIDNameMap(kt *kit.Kit, resTypeIDsMap map[client.TypeID][]string) (map[string]string, error) {

	nameMap := make(map[string]string)
	for resType, ids := range resTypeIDsMap {
		switch resType {
		case sys.Business:
			for _, id := range ids {
				space, err := a.spaceMgr.GetSpaceByUID(id)
				if err != nil {
					return nil, err
				}
				nameMap[id] = space.SpaceName
			}
		case sys.Application:
			for _, id := range ids {
				i, err := strconv.Atoi(id)
				if err != nil {
					return nil, err
				}
				app, err := a.ds.GetAppByID(kt.Ctx, &pbds.GetAppByIDReq{AppId: uint32(i)})
				if err != nil {
					return nil, err
				}
				nameMap[id] = app.Spec.Name
			}
		case sys.AppCredential:
			return nil, fmt.Errorf("NOT IMPLEMENTED")
		}
	}
	return nameMap, nil
}

// GrantResourceCreatorAction grant resource creator action.
func (a *Auth) GrantResourceCreatorAction(ctx context.Context, opts *client.GrantResourceCreatorActionOption) error {
	return a.auth.GrantResourceCreatorAction(ctx, opts)
}
