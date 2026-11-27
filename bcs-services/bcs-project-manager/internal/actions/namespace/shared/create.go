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

package shared

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace/independent"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bcscc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	v2 "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/itsm/v2"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vvm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	quotautils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/quota"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// NamespacePrefix namespace in shared cluster must be prefixed by it
var NamespacePrefix = envs.BCSNamespacePrefix + "-%s-"

// CreateNamespace implement for CreateNamespace interface
func (a *SharedNamespaceAction) CreateNamespace(ctx context.Context,
	req *proto.CreateNamespaceRequest, resp *proto.CreateNamespaceResponse) error {
	// 共享集群管理页创建：显式指定归属项目时，由集群所属项目免审批直建并写入归属
	if req.GetTargetProjectCode() != "" {
		return a.createWithTargetProject(ctx, req)
	}
	if err := a.validateCreate(ctx, req); err != nil {
		return err
	}
	// if itsm is not enable, create namespace directly
	if !config.GlobalConf.ITSM.Enable {
		ia := independent.NewIndependentNamespaceAction(a.model)
		req.Annotations = append(req.Annotations, &proto.Annotation{
			Key:   config.GlobalConf.SharedClusterConfig.AnnoKeyProjCode,
			Value: req.GetProjectCode(),
		})
		return ia.CreateNamespace(ctx, req, resp)
	}
	var username string
	if authUser, gErr := middleware.GetUserFromContext(ctx); gErr == nil {
		username = authUser.GetUsername()
	}
	namespace := &nsm.Namespace{
		ProjectCode: req.GetProjectCode(),
		ClusterID:   req.GetClusterID(),
		Name:        req.GetName(),
		Creator:     username,
		Updater:     username,
		Managers:    username,
		ResourceQuota: &nsm.Quota{
			CPURequests:    req.GetQuota().GetCpuRequests(),
			CPULimits:      req.GetQuota().GetCpuLimits(),
			MemoryRequests: req.GetQuota().GetMemoryRequests(),
			MemoryLimits:   req.GetQuota().GetMemoryLimits(),
		},
		UpdateTime: time.Now().UTC().Format(time.RFC3339),
	}
	variables := []*nsm.Variable{}
	for _, variable := range req.GetVariables() {
		variables = append(variables, &nsm.Variable{
			VariableID: variable.Id,
			ClusterID:  variable.ClusterID,
			Namespace:  variable.Namespace,
			Key:        variable.Key,
			Value:      variable.Value,
		})
	}
	namespace.Variables = variables
	cpuLimits, err := resource.ParseQuantity(req.GetQuota().GetCpuLimits())
	if err != nil {
		logging.Error("parse quantity cpu limits failed, err: %s", err.Error())
		return err
	}
	memoryLimits, err := resource.ParseQuantity(req.GetQuota().GetMemoryLimits())
	if err != nil {
		logging.Error("parse quantity cpu limits failed, err: %s", err.Error())
		return err
	}
	// memoryLimits.Value() return unit is byte， needs to be converted to Gi (divide 2^30)
	// TODO ITSM 传入 tenantID
	itsmResp, err := v2.SubmitCreateNamespaceTicket(ctx, username, req.GetProjectCode(), req.GetClusterID(), req.GetName(),
		int(cpuLimits.Value()), int(memoryLimits.Value()/int64(math.Pow(2, 30))))
	if err != nil {
		logging.Error("itsm create ticket failed, err: %s", err.Error())
		return err
	}
	namespace.ItsmTicketType = nsm.ItsmTicketTypeCreate
	namespace.ItsmTicketURL = itsmResp.TicketURL
	namespace.ItsmTicketStatus = nsm.ItsmTicketStatusCreated
	namespace.ItsmTicketSN = itsmResp.SN
	if err := a.model.CreateNamespace(ctx, namespace); err != nil {
		logging.Error("create namespace %s/%s in db failed, err: %s", req.GetClusterID(), req.GetName(), err.Error())
		return err
	}
	return nil
}

func (a *SharedNamespaceAction) validateCreate(ctx context.Context, req *proto.CreateNamespaceRequest) error {
	// 归属项目必须在集群可见范围内（集群所属项目自身可直接创建）
	cluster, err := clustermanager.GetCluster(ctx, req.ClusterID, true)
	if err != nil {
		logging.Error("get cluster %s failed, err: %s", req.ClusterID, err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	ownerProject, err := a.model.GetProject(ctx, cluster.GetProjectID())
	if err != nil {
		logging.Error("get cluster owner project %s failed, err: %s", cluster.GetProjectID(), err.Error())
		return errorx.NewDBErr(err.Error())
	}
	if req.ProjectCode != ownerProject.ProjectCode {
		targetProject, tErr := a.model.GetProject(ctx, req.ProjectCode)
		if tErr != nil {
			return errorx.NewReadableErr(errorx.ProjectNotExistsErr, "归属项目不存在")
		}
		if err := a.checkProjectInSharedRanges(cluster, targetProject); err != nil {
			return err
		}
	}
	// check is namespace name valid
	return a.validateNamespaceName(ctx, req, req.ProjectCode)
}

// validateNamespaceName 校验命名空间名称（前缀按指定归属项目编码）、唯一性与配额
func (a *SharedNamespaceAction) validateNamespaceName(ctx context.Context,
	req *proto.CreateNamespaceRequest, projectCode string) error {
	// 长名回退：bcs-{projectCode}- 前缀 + 自定义部分超 63 字符时，允许使用平台前缀 bcs-
	fullPrefix := fmt.Sprintf(NamespacePrefix, projectCode)
	platformPrefix := envs.BCSNamespacePrefix + "-"
	if !strings.HasPrefix(req.Name, fullPrefix) {
		customPart := strings.TrimPrefix(req.Name, platformPrefix)
		wouldExceed := len(fullPrefix)+len(customPart) > 63
		if !(strings.HasPrefix(req.Name, platformPrefix) && wouldExceed) {
			return errorx.NewReadableErr(errorx.ParamErr, fmt.Sprintf("共享集群命名空间必须以 %s-[projectCode]- 开头",
				envs.BCSNamespacePrefix))
		}
	}
	// check is namespace name exists
	stagings, _ := a.model.ListNamespacesByItsmTicketType(ctx,
		projectCode, req.ClusterID, []string{nsm.ItsmTicketTypeCreate})
	for _, staging := range stagings {
		if staging.Name == req.Name {
			return errorx.NewReadableErr(errorx.ParamErr, fmt.Sprintf("命名空间 [%s] 已存在", req.Name))
		}
	}
	client, err := clientset.GetClientGroup().Client(req.ClusterID)
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.ClusterID, err.Error())
		return err
	}
	_, err = client.CoreV1().Namespaces().Get(ctx, req.Name, metav1.GetOptions{})
	if err == nil {
		return errorx.NewReadableErr(errorx.ParamErr, fmt.Sprintf("命名空间 [%s] 已存在", req.Name))
	}
	if !errors.IsNotFound(err) {
		logging.Error("get namespace in cluster %s failed, err: %s", req.ClusterID, err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	// 创建/分配时强制设置 CPU/内存配额，共享集群创建不允许缺省配额
	if req.GetQuota() == nil {
		return errorx.NewReadableErr(errorx.ParamErr, "命名空间配额不能为空")
	}
	// check resourceQuota validate
	if err := quotautils.ValidateResourceQuota(req.Quota); err != nil {
		return err
	}
	if !config.GlobalConf.SharedClusterConfig.AllowAdjustQuota {
		if err := quotautils.ValidateQuotaEquality(req.Quota); err != nil {
			return err
		}
	}
	return nil
}

// createWithTargetProject 共享集群管理页创建：由集群所属项目发起，免审批直建，
// 命名空间归属由 targetProjectCode 指定（须在集群可见范围内），不自动授予创建者命名空间权限
func (a *SharedNamespaceAction) createWithTargetProject(ctx context.Context,
	req *proto.CreateNamespaceRequest) error {
	cluster, err := clustermanager.GetCluster(ctx, req.ClusterID, true)
	if err != nil {
		logging.Error("get cluster %s failed, err: %s", req.ClusterID, err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	if !cluster.GetIsShared() {
		return errorx.NewReadableErr(errorx.ParamErr, "非共享集群不支持指定归属项目创建命名空间")
	}
	// 该入口为集群所属项目专用（共享管理页），防止使用方项目借直建通道绕过审批
	ownerProject, ownerErr := a.model.GetProject(ctx, cluster.GetProjectID())
	if ownerErr != nil {
		logging.Error("get cluster owner project %s failed, err: %s", cluster.GetProjectID(), ownerErr.Error())
		return errorx.NewDBErr(ownerErr.Error())
	}
	if req.GetProjectCode() != ownerProject.ProjectCode {
		return errorx.NewReadableErr(errorx.ParamErr, "仅集群所属项目支持指定归属项目创建命名空间")
	}
	targetProject, tErr := a.model.GetProject(ctx, req.GetTargetProjectCode())
	if tErr != nil {
		return errorx.NewReadableErr(errorx.ProjectNotExistsErr, "归属项目不存在")
	}
	if err := a.checkProjectInSharedRanges(cluster, targetProject); err != nil {
		return err
	}
	// 名称前缀与唯一性、配额按归属项目校验
	if err := a.validateNamespaceName(ctx, req, req.GetTargetProjectCode()); err != nil {
		return err
	}
	return a.directCreateNamespace(ctx, req, targetProject.ProjectCode)
}

// directCreateNamespace 免审批直建：写归属注解与配额，不授予创建者命名空间权限，不走 ITSM
func (a *SharedNamespaceAction) directCreateNamespace(ctx context.Context,
	req *proto.CreateNamespaceRequest, targetProjectCode string) error {
	var creator string
	if authUser, gErr := middleware.GetUserFromContext(ctx); gErr == nil {
		creator = authUser.GetUsername()
	}
	client, err := clientset.GetClientGroup().Client(req.ClusterID)
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.ClusterID, err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	ns := &corev1.Namespace{}
	ns.SetName(req.GetName())
	labels := map[string]string{}
	for _, label := range req.GetLabels() {
		labels[label.GetKey()] = label.GetValue()
	}
	ns.SetLabels(labels)
	annotations := map[string]string{
		config.GlobalConf.SharedClusterConfig.AnnoKeyProjCode: targetProjectCode,
	}
	if creator != "" {
		annotations[constant.AnnotationKeyCreator] = creator
	}
	ns.SetAnnotations(annotations)
	if _, cErr := client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{}); cErr != nil {
		logging.Error("create namespace %s in cluster %s failed, err: %s",
			req.GetName(), req.GetClusterID(), cErr.Error())
		return errorx.NewClusterErr(cErr.Error())
	}
	// 创建完成后不自动授予创建者命名空间权限
	// create resource quota
	if req.GetQuota() != nil {
		quota := &corev1.ResourceQuota{
			Spec: corev1.ResourceQuotaSpec{
				Hard: corev1.ResourceList{},
			},
		}
		quota.SetName(req.GetName())
		quota.SetNamespace(req.GetName())
		if lErr := quotautils.LoadFromProto(quota, req.GetQuota()); lErr != nil {
			logging.Error("load quota from proto for %s failed, err: %s", req.GetName(), lErr.Error())
			return lErr
		}
		if _, qErr := client.CoreV1().ResourceQuotas(req.GetName()).Create(ctx, quota, metav1.CreateOptions{}); qErr != nil {
			logging.Error("create resourcequota %s in cluster %s failed, err: %s",
				req.GetName(), req.GetClusterID(), qErr.Error())
			return errorx.NewClusterErr(qErr.Error())
		}
	}
	// create variables
	for _, variable := range req.GetVariables() {
		entity := &vvm.VariableValue{
			VariableID: variable.GetId(),
			Value:      variable.GetValue(),
			ClusterID:  req.GetClusterID(),
			Namespace:  req.GetName(),
			Scope:      vdm.VariableScopeNamespace,
		}
		if uErr := a.model.UpsertVariableValue(ctx, entity); uErr != nil {
			logging.Error("create variable in %s/%s failed, err: %s", req.GetClusterID(), req.GetName(), uErr.Error())
			return errorx.NewDBErr(uErr.Error())
		}
	}
	// sync to bcs-cc
	go func() {
		if err := bcscc.CreateNamespace(targetProjectCode, req.GetClusterID(), req.GetName(), creator); err != nil {
			logging.Error("[ALARM-CC-NAMESPACE] create namespace %s/%s/%s in paas-cc failed, err: %s",
				targetProjectCode, req.GetClusterID(), req.GetName(), err.Error())
		}
	}()
	return nil
}
