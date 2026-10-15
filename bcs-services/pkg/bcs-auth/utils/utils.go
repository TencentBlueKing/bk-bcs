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

// Package utils xxx
package utils

import (
	// NOCC:gas/crypto(设计如此)
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	bkiam "github.com/TencentBlueKing/iam-go-sdk"
	blog "k8s.io/klog/v2"
)

var (
	// ErrServerNotInited server not init
	ErrServerNotInited = errors.New("server not init")
)

const (
	// PermDeniedCode perm denied code
	PermDeniedCode = 40300
	// DefaultTenantId default tenant id
	DefaultTenantId = "default"
)

// UserInfo user info for perm check
type UserInfo struct {
	BkUserName string
	TenantId   string
}

// GetBKUserName get bk user name
func (u UserInfo) GetBKUserName() string {
	return u.BkUserName
}

// GetTenantId get tenant id
func (u UserInfo) GetTenantId() string {
	if u.TenantId == "" {
		return DefaultTenantId
	}
	return u.TenantId
}

// String xxx
func (u UserInfo) String() string {
	return fmt.Sprintf("TenantId: %s, BkUserName: %s", u.GetTenantId(), u.GetBKUserName())
}

// PermDeniedError permission denied,user need to apply
type PermDeniedError struct {
	Perms PermData `json:"perms"`
}

// PermData permission data for no permission
type PermData struct {
	ApplyURL   string           `json:"apply_url"`
	ActionList []ResourceAction `json:"action_list"`
}

// Error return error message with perm actions
func (e *PermDeniedError) Error() string {
	var actions string
	for _, action := range e.Perms.ActionList {
		actions = actions + " " + action.Action
	}
	return fmt.Sprintf("permission denied, need%s permition", actions)
}

// ResourceAction for multi action multi resources
type ResourceAction struct {
	Resource string `json:"-"`
	Type     string `json:"resource_type"`
	Action   string `json:"action_id"`
}

// CheckResourceRequest xxx
type CheckResourceRequest struct {
	Module    string
	Operation string
	User      string
}

// CheckResourcePerms check multi resources actions in perms
func CheckResourcePerms(req CheckResourceRequest, resources []ResourceAction,
	perms map[string]map[string]bool) (bool, error) {
	if len(perms) == 0 {
		return false, fmt.Errorf("checkResourcePerms get perm empty")
	}

	for _, r := range resources {
		perm, ok := perms[r.Resource]
		if !ok {
			blog.Errorf("%s %s user[%s] resource[%s] not exist in perms", req.Module,
				req.Operation, req.User, r.Resource)
			return false, nil
		}

		if !perm[r.Action] {
			blog.Infof("%s %s user[%s] resource[%v] action[%s] allow[%v]",
				req.Module, req.Operation, req.User, r.Resource, r.Action, perm[r.Action])
			return false, nil
		}
	}

	return true, nil
}

// https://bk.tencent.com/docs/document/6.0/160/8462

// ClusterApplication build iam.Application for ActionID
type ClusterApplication struct {
	ActionID string
}

// BuildIAMApplication only support same system same cluster
func BuildIAMApplication(app ClusterApplication,
	resourceTypes []bkiam.ApplicationRelatedResourceType) iam.ApplicationAction {
	applicationAction := iam.ApplicationAction{
		ActionID:         app.ActionID,
		RelatedResources: make([]bkiam.ApplicationRelatedResourceType, 0),
	}
	if len(resourceTypes) > 0 {
		applicationAction.RelatedResources = append(applicationAction.RelatedResources, resourceTypes...)
	}

	return applicationAction
}

// BuildRelatedSystemResource build application related resourceInstance
func BuildRelatedSystemResource(systemID, resourceType string,
	instances [][]iam.Instance) bkiam.ApplicationRelatedResourceType {
	relatedResource := bkiam.ApplicationRelatedResourceType{
		SystemID:  systemID,
		Type:      resourceType,
		Instances: make([]bkiam.ApplicationResourceInstance, 0),
	}
	if len(instances) > 0 {
		for i := range instances {
			relatedResource.Instances = append(relatedResource.Instances, iam.BuildResourceInstance(instances[i]))
		}
	}

	return relatedResource
}

// ResourceInfo resource
type ResourceInfo struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Validate validate
func (rs ResourceInfo) Validate() error {
	if rs.ID == "" || rs.Type == "" || rs.Name == "" {
		return fmt.Errorf("ResourceInfo resource empty")
	}

	return nil
}

// AuthorizeCreatorOptions authorize creator options
type AuthorizeCreatorOptions struct {
	Ancestors []iam.Ancestor
}

// AuthorizeCreatorOption xxx
type AuthorizeCreatorOption func(options *AuthorizeCreatorOptions)

// WithAncestors set authorizeCreatorOptions ancestors
func WithAncestors(ancestors []iam.Ancestor) AuthorizeCreatorOption {
	return func(q *AuthorizeCreatorOptions) {
		q.Ancestors = ancestors
	}
}

// CalcIAMNsID 计算命名空间在 IAM 中的 ID，格式：{集群5位数字ID}:{md5(命名空间名称)}{命名空间名称前两位}
// 如 `BCS-K8S-40000:default` 会被处理成 `40000:5f03d33dde`
func CalcIAMNsID(clusterID, namespace string) string {
	s := strings.Split(clusterID, "-")
	clusterIDNum := s[len(s)-1]
	// NOCC:gas/crypto(设计如此)
	// nolint
	h := md5.New()
	_, _ = io.WriteString(h, namespace)
	b := h.Sum(nil)
	name := namespace
	if len(namespace) >= 2 {
		name = namespace[:2]
	}
	return fmt.Sprintf("%s:%x%s", clusterIDNum, b[4:8], name)
}

// GetEnvWithDefault takes two string parameters, key and defaultValue.
// It uses the os.Getenv function to retrieve the value of the environment variable specified by key.
// If the value is an empty string, it returns the defaultValue parameter.
// Otherwise, it returns the value of the environment variable.
func GetEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GenerateEventID generate event id, format: app_code-YYYYMMDDHHMMSS-substring(MD5(随机因子)),8,24)
func GenerateEventID(appCode, factor string) string {
	currentTime := time.Now().Format("20060102150405")
	// NOCC:gas/crypto(设计如此)
	// nolint
	hash := fmt.Sprintf("%x", md5.Sum([]byte(factor)))
	result := fmt.Sprintf("%s-%s-%s", appCode, currentTime, hash[8:24])
	return result
}
