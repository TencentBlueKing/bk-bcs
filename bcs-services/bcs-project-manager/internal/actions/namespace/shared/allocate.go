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
	"time"

	cms "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	quotautils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/quota"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// AllocateNamespace implement for AllocateNamespace interface
// 语义：原归属为空 → 分配（写 DB 记录 + 写归属 annotation + 下发 ResourceQuota）；
// 原归属非空 → 变更归属（更新归属项目 + annotation + 重新下发配额）
func (a *SharedNamespaceAction) AllocateNamespace(ctx context.Context,
	req *proto.AllocateNamespaceRequest, resp *proto.AllocateNamespaceResponse) error {
	if err := a.validateAllocate(ctx, req); err != nil {
		return err
	}
	cluster, err := clustermanager.GetCluster(ctx, req.GetClusterID(), true)
	if err != nil {
		logging.Error("get cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	// 仅集群所属项目可分配/变更归属（限制:共享管理页仅所属项目可见）
	operatorProject, err := a.model.GetProject(ctx, req.GetProjectCode())
	if err != nil {
		logging.Error("get operator project %s failed, err: %s", req.GetProjectCode(), err.Error())
		return errorx.NewReadableErr(errorx.ProjectNotExistsErr, "操作方项目不存在")
	}
	if operatorProject.ProjectID != cluster.GetProjectID() {
		return errorx.NewReadableErr(errorx.PermDeniedErr, "仅集群所属项目可分配命名空间")
	}
	// 目标归属项目必须存在且在集群可见范围内
	targetProject, err := a.model.GetProject(ctx, req.GetTargetProjectCode())
	if err != nil {
		logging.Error("get target project %s failed, err: %s", req.GetTargetProjectCode(), err.Error())
		return errorx.NewReadableErr(errorx.ProjectNotExistsErr, "目标归属项目不存在")
	}
	if err := a.checkProjectInSharedRanges(cluster, targetProject); err != nil {
		return err
	}

	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	ns, err := client.CoreV1().Namespaces().Get(ctx, req.GetNamespace(), metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return errorx.NewReadableErr(errorx.ParamErr, fmt.Sprintf("命名空间 [%s] 不存在", req.GetNamespace()))
	}
	if err != nil {
		logging.Error("get namespace %s in cluster %s failed, err: %s",
			req.GetNamespace(), req.GetClusterID(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	// vcluster 占用的命名空间不可分配
	if _, ok := ns.GetAnnotations()[constant.AnnotationKeyVcluster]; ok {
		return errorx.NewReadableErr(errorx.ParamErr, "vcluster 占用的命名空间，不可分配")
	}

	// 配额保留策略：未指定配额时，已有 ResourceQuota 的命名空间保留原配额不覆盖； 无配额的命名空间必须显式指定配额
	existingQuota, quotaErr := client.CoreV1().ResourceQuotas(req.GetNamespace()).
		Get(ctx, req.GetNamespace(), metav1.GetOptions{})
	if quotaErr != nil && !k8serrors.IsNotFound(quotaErr) {
		logging.Error("get resourcequota %s/%s failed, err: %s",
			req.GetClusterID(), req.GetNamespace(), quotaErr.Error())
		return errorx.NewClusterErr(quotaErr.Error())
	}
	hasExistingQuota := quotaErr == nil
	if req.GetQuota() == nil && !hasExistingQuota {
		return errorx.NewReadableErr(errorx.ParamErr,
			fmt.Sprintf("命名空间 [%s] 无配额，分配时必须设置配额", req.GetNamespace()))
	}
	// 批量分配语义（N6 备注）：默认（overwrite=false）时请求配额仅对无ResourceQuota的
	// overwrite=true 时显式覆盖，供单个手动调整场景使用
	effectiveQuota := req.GetQuota()
	if req.GetQuota() != nil && hasExistingQuota && !req.GetOverwrite() {
		effectiveQuota = nil
	}

	annoKey := config.GlobalConf.SharedClusterConfig.AnnoKeyProjCode
	originProjectCode := ns.GetAnnotations()[annoKey]
	// 审批中的命名空间不可分配/变更归属
	if originProjectCode != "" {
		if err := a.checkPendingTicket(ctx, originProjectCode, req.GetClusterID(), req.GetNamespace()); err != nil {
			return err
		}
	}
	if err := a.checkPendingTicket(ctx, targetProject.ProjectCode, req.GetClusterID(), req.GetNamespace()); err != nil {
		return err
	}

	var username string
	if authUser, gErr := middleware.GetUserFromContext(ctx); gErr == nil {
		username = authUser.GetUsername()
	}

	// 写归属 annotation（分配/变更归属）
	if ns.GetAnnotations() == nil {
		ns.Annotations = map[string]string{}
	}
	ns.Annotations[annoKey] = targetProject.ProjectCode
	if _, err := client.CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{}); err != nil {
		logging.Error("update namespace %s/%s annotation failed, err: %s",
			req.GetClusterID(), req.GetNamespace(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	// 下发 ResourceQuota（使用最终生效配额：保留场景为 nil，跳过下发）
	if err := applyResourceQuota(ctx, req.GetClusterID(), req.GetNamespace(), effectiveQuota); err != nil {
		return err
	}

	// 写入DB记录：分配（无归属记录）新建；变更归属转移；归属不变仅更新配额；配额取值：优先请求指定值；未指定时保留k8s现有 ResourceQuota 配额
	quotaModel := resolveQuotaModel(effectiveQuota, existingQuota)
	now := time.Now().UTC().Format(time.RFC3339)
	createRecord := func(projectCode string) error {
		nsModel := &nsm.Namespace{
			ProjectCode:   projectCode,
			ClusterID:     req.GetClusterID(),
			Name:          req.GetNamespace(),
			Creator:       username,
			Updater:       username,
			ResourceQuota: quotaModel,
			UpdateTime:    now,
		}
		if err := a.model.CreateNamespace(ctx, nsModel); err != nil {
			logging.Error("create namespace %s/%s in db failed, err: %s",
				req.GetClusterID(), req.GetNamespace(), err.Error())
			return errorx.NewDBErr(err.Error())
		}
		return nil
	}
	if originProjectCode == "" {
		if err := createRecord(targetProject.ProjectCode); err != nil {
			return err
		}
	} else if originProjectCode != targetProject.ProjectCode {
		// 变更归属
		if _, gErr := a.model.GetNamespace(ctx, originProjectCode, req.GetClusterID(), req.GetNamespace()); gErr != nil {
			// annotation 有归属但无 DB 记录，按分配处理新建记录
			if err := createRecord(targetProject.ProjectCode); err != nil {
				return err
			}
		} else if _, err := a.model.TransferNamespace(ctx, originProjectCode, targetProject.ProjectCode,
			req.GetClusterID(), req.GetNamespace(), quotaModel, username); err != nil {
			logging.Error("transfer namespace %s/%s from %s to %s failed, err: %s",
				req.GetClusterID(), req.GetNamespace(), originProjectCode,
				targetProject.ProjectCode, err.Error())
			return errorx.NewDBErr(err.Error())
		}
	} else {
		// 归属不变，仅更新配额与更新时间
		if _, err := a.model.UpdateNamespace(ctx, entity.M{
			nsm.FieldKeyProjectCode: originProjectCode,
			nsm.FieldKeyClusterID:   req.GetClusterID(),
			nsm.FieldKeyName:        req.GetNamespace(),
			"resourceQuota":         quotaModel,
			"updater":               username,
			nsm.FieldKeyUpdateTime:  now,
		}); err != nil {
			logging.Error("update namespace %s/%s quota in db failed, err: %s",
				req.GetClusterID(), req.GetNamespace(), err.Error())
			return errorx.NewDBErr(err.Error())
		}
	}

	resp.Data = &proto.NamespaceData{
		Name: req.GetNamespace(),
		Quota: &proto.ResourceQuota{
			CpuRequests:    quotaModel.CPURequests,
			CpuLimits:      quotaModel.CPULimits,
			MemoryRequests: quotaModel.MemoryRequests,
			MemoryLimits:   quotaModel.MemoryLimits,
		},
		CreateTime: ns.GetCreationTimestamp().UTC().Format(time.RFC3339),
	}
	return nil
}

// validateAllocate 校验分配请求的基础参数
// Q4：配额可选——未指定时已有配额的命名空间保留原配额；指定时校验格式与 request = limit
func (a *SharedNamespaceAction) validateAllocate(ctx context.Context, req *proto.AllocateNamespaceRequest) error {
	if req.GetQuota() != nil {
		if err := quotautils.ValidateResourceQuota(req.Quota); err != nil {
			return err
		}
		if !config.GlobalConf.SharedClusterConfig.AllowAdjustQuota {
			if err := quotautils.ValidateQuotaEquality(req.Quota); err != nil {
				return err
			}
		}
	}
	if req.GetTargetProjectCode() == "" {
		return errorx.NewReadableErr(errorx.ParamErr, "目标归属项目编码不能为空")
	}
	return nil
}

// checkProjectInSharedRanges 校验目标项目是否在集群共享可见范围内
// sharedRanges 为空表示全平台共享；仅含 Bizs 表示按业务共享，不做项目级限制
func (a *SharedNamespaceAction) checkProjectInSharedRanges(cluster *cms.Cluster, target *pm.Project) error {
	ranges := cluster.GetSharedRanges()
	if ranges == nil {
		return nil
	}
	idOrCodes := ranges.GetProjectIdOrCodes()
	if len(idOrCodes) == 0 {
		return nil
	}
	for _, idOrCode := range idOrCodes {
		if idOrCode == target.ProjectCode || idOrCode == target.ProjectID {
			return nil
		}
	}
	return errorx.NewReadableErr(errorx.ParamErr,
		fmt.Sprintf("项目 [%s] 不在集群可见范围内", target.ProjectCode))
}

// checkPendingTicket 校验命名空间是否存在进行中的审批单
func (a *SharedNamespaceAction) checkPendingTicket(ctx context.Context,
	projectCode, clusterID, namespace string) error {
	record, err := a.model.GetNamespace(ctx, projectCode, clusterID, namespace)
	if err != nil {
		// 无 DB 记录，视为无进行中审批
		return nil
	}
	if record.ItsmTicketStatus == nsm.ItsmTicketStatusCreated {
		return errorx.NewReadableErr(errorx.ParamErr,
			fmt.Sprintf("命名空间 [%s] 存在进行中的审批单，暂不可操作", namespace))
	}
	return nil
}

// resolveQuotaModel 优先使用请求指定配额；未指定时保留 k8s 现有 ResourceQuota 配额
func resolveQuotaModel(reqQuota *proto.ResourceQuota, existing *corev1.ResourceQuota) *nsm.Quota {
	if reqQuota != nil {
		return &nsm.Quota{
			CPURequests:    reqQuota.GetCpuRequests(),
			CPULimits:      reqQuota.GetCpuLimits(),
			MemoryRequests: reqQuota.GetMemoryRequests(),
			MemoryLimits:   reqQuota.GetMemoryLimits(),
		}
	}
	hard := existing.Spec.Hard
	return &nsm.Quota{
		CPURequests:    quotaSpecString(hard, corev1.ResourceRequestsCPU),
		CPULimits:      quotaSpecString(hard, corev1.ResourceLimitsCPU),
		MemoryRequests: quotaSpecString(hard, corev1.ResourceRequestsMemory),
		MemoryLimits:   quotaSpecString(hard, corev1.ResourceLimitsMemory),
	}
}

// quotaSpecString 取 ResourceList 中指定资源的配额值，缺失返回空串
func quotaSpecString(hard corev1.ResourceList, key corev1.ResourceName) string {
	if q, ok := hard[key]; ok {
		return q.String()
	}
	return ""
}

// applyResourceQuota 在命名空间上创建或更新 ResourceQuota
// quota 为 nil 时保留现有配额不覆盖（调用前需保证已有 ResourceQuota，否则无配额可保留）
func applyResourceQuota(ctx context.Context, clusterID, namespace string, quota *proto.ResourceQuota) error {
	if quota == nil {
		return nil
	}
	client, err := clientset.GetClientGroup().Client(clusterID)
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", clusterID, err.Error())
		return err
	}
	k8sQuota := &corev1.ResourceQuota{
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{},
		},
	}
	k8sQuota.SetName(namespace)
	k8sQuota.SetNamespace(namespace)
	if err := quotautils.LoadFromProto(k8sQuota, quota); err != nil {
		logging.Error("load resource quota from proto failed, err: %s", err.Error())
		return errorx.NewParamErr("invalid quota")
	}
	existing, err := client.CoreV1().ResourceQuotas(namespace).Get(ctx, namespace, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		if _, cErr := client.CoreV1().ResourceQuotas(namespace).Create(ctx, k8sQuota, metav1.CreateOptions{}); cErr != nil {
			logging.Error("create resourcequota %s/%s failed, err: %s", clusterID, namespace, cErr.Error())
			return errorx.NewClusterErr(cErr.Error())
		}
		return nil
	}
	if err != nil {
		logging.Error("get resourcequota %s/%s failed, err: %s", clusterID, namespace, err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	existing.Spec.Hard = k8sQuota.Spec.Hard
	if _, uErr := client.CoreV1().ResourceQuotas(namespace).Update(ctx, existing, metav1.UpdateOptions{}); uErr != nil {
		logging.Error("update resourcequota %s/%s failed, err: %s", clusterID, namespace, uErr.Error())
		return errorx.NewClusterErr(uErr.Error())
	}
	return nil
}
