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

// Package core xxx
package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/repository"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/runner"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/utils"
)

// Task 处理Reconcile核心逻辑，避免Reconcile方法复杂化
type Task interface {
	// Init tf cr
	Init() error
	// CheckForChanges 检查commit-id是否发生变化
	CheckForChanges() bool
	// Plan 执行plan
	Plan() (string, error)
	// SaveTfPlan 保存plan结果
	SaveTfPlan() error
	// GetTfPlan 从api server获取tf plan
	GetTfPlan() (string, error)
	// CheckCommitIdConsistency 检查commit-id一致性
	CheckCommitIdConsistency(currentCommitId string) bool
	// Apply tf cr
	Apply() (string, error)
	// SaveApplyOutputToConfigMap 保存apply输出结果
	SaveApplyOutputToConfigMap(raw string) error
	// GetLastCommitId 获取分支最后一次commit-id
	GetLastCommitId() (string, error)
}

// NewTask new Task
func NewTask(rootCtx context.Context, traceId string, terraform *tfv1.Terraform, client client.Client) Task {
	ctx, cancel := context.WithCancel(rootCtx)
	return &task{
		rootCtx:   ctx,
		cancel:    cancel,
		client:    client,
		traceId:   traceId,
		terraform: terraform,
		nn: types.NamespacedName{
			Name:      terraform.Name,
			Namespace: terraform.Namespace,
		},
	}
}

// task impl Task
type task struct {
	// traceId 链路id
	traceId string
	// rootCtx root 上下文
	rootCtx context.Context
	// cancel 取消函数
	cancel context.CancelFunc

	// nn 名称空间和名称
	nn types.NamespacedName
	// terraform 请求对象
	terraform *tfv1.Terraform
	// client api server client
	client client.Client

	// repoToken token
	repoToken string
	// repo repository obj
	repo repository.Handler
	// lastCommitId last commit-id
	lastCommitId string

	// runner tf local runner
	runner runner.LocalTerraform
}

// Init 拉取repo、创建tf命令行工具
func (h *task) Init() error {
	// 1. get terraform source, i.e. terraform code in git
	if err := h.getRepoToken(); err != nil {
		return err
	}

	h.repo = repository.NewHandler(h.rootCtx, h.terraform.Spec.Repository, h.repoToken, h.traceId, h.nn)
	if err := h.repo.Init(); err != nil {
		return errors.Wrapf(err, "repository init failed")
	}

	return nil
}

// CheckForChanges 检查commit-id是否发生变化
func (h *task) CheckForChanges() bool {
	if len(h.lastCommitId) == 0 {
		return false
	}

	revision := h.terraform.Status.LastAppliedRevision
	if len(revision) == 0 { // 第一次，设为有变化
		return true
	}

	if revision == h.lastCommitId || revision == h.lastCommitId[:len(revision)] { // length: 8位或者9位
		// 无变化
		return false
	}

	// 有变化
	return true
}

// CheckCommitIdConsistency 检查commit-id一致性
func (h *task) CheckCommitIdConsistency(currentCommitId string) bool {
	length := len(currentCommitId)

	if currentCommitId == h.lastCommitId || (length >= 8 && currentCommitId == h.lastCommitId[:length]) { // 8位及以上
		return true
	}

	return false
}

// GetTfPlan 从api server获取tf plan
// 1.从api server查询tfplan
// 2.检查tfplan数据是否存在
// 3.解压tfplan的数据
// 4.创建目录、文件
// 5.tfplan写回磁盘
// 6.创建terraform
// 7.初始化terraform
func (h *task) GetTfPlan() (string, error) {
	if err := h.repo.Pull(); err != nil {
		return "", errors.Wrapf(err, "pull remote repository failed")
	}

	tfplanSecret := new(corev1.Secret)
	secretName := fmt.Sprintf("tfplan-%s", h.terraform.Name)
	tfplanObjectKey := types.NamespacedName{Name: secretName, Namespace: h.terraform.Namespace}

	ctx, cancel := context.WithTimeout(h.rootCtx, 15*time.Second)
	defer cancel()
	if err := h.client.Get(ctx, tfplanObjectKey, tfplanSecret); err != nil {
		return "", errors.Wrapf(err, "unable to get the plan secret, tf: %s, trace-id: %s", h.nn, h.traceId)
	}

	bs, ok := tfplanSecret.Data[TFPlanName]
	if !ok || len(bs) == 0 {
		return "", errors.Errorf("unable to obtain tfplan data from the secret, 'tfplan' key does not exist, "+
			"tf: %s, trace-id: %s", h.nn, h.traceId)
	}
	blog.Infof("get tfplan result success, tfplan: %s, tf: %s, trace-id: %s", secretName, h.nn, h.traceId)

	tfplan, err := utils.GzipDecode(bs)
	if err != nil {
		return "", errors.Wrapf(err, "gzip decode failed, tf: %s, trace-id: %s", h.nn, h.traceId)
	}

	fileFullPath := filepath.Join(h.repo.GetExecutePath(), fmt.Sprintf("main-%s.tfplan", h.traceId))
	if err = afero.WriteFile(afero.NewOsFs(), fileFullPath, tfplan, 0644); err != nil {
		return "", errors.Wrapf(err, "unable to write the plan to disk")
	}
	blog.Info("writer tfplan result success, tfplan: %s, tf: %s, trace-id: %s", fileFullPath, h.nn, h.traceId)

	req := &runner.NewTerraformRequest{
		WorkingDir: h.repo.GetExecutePath(),
		InstanceID: h.traceId,
		Terraform:  h.terraform,
	}
	tfRunner, err := runner.NewLocalTerraform(h.client, h.terraform.Spec.Project)
	if err != nil {
		return "", errors.Wrapf(err, "new local terraform failed, terraform: %s, trace-id: %s", h.nn, h.traceId)
	}
	h.runner = tfRunner
	if _, err = h.runner.Init(req); err != nil {
		return "", errors.Wrapf(err, "new terraform failed, terraform: %s, trace-id: %s", h.nn, h.traceId)
	}
	h.runner.SetPlanOutFile(fileFullPath)

	/* init */
	_, err = h.runner.ExecInit(h.rootCtx, &runner.InitRequest{Upgrade: false, ForceCopy: true, TfInstance: h.traceId})
	logs := h.runner.GetInitLog()
	if err != nil {
		return logs, errors.Wrapf(err, "init terraform failed, terraform: %s, trace-id: %s", h.nn, h.traceId)
	}
	blog.Info("init terraform runner success, terraform: %s, trace-id: %s", h.nn, h.traceId)

	return logs, nil
}

// Plan exec plan
// 1.获取代码
// 2.创建tf命令行工具
// 3.初始化tf命令行工具
// 4.执行init命令
// 5.执行plan命令
func (h *task) Plan() (string, error) {
	if err := h.repo.Pull(); err != nil {
		return "", errors.Wrapf(err, "pull remote repository failed")
	}
	req := &runner.NewTerraformRequest{
		InstanceID: h.traceId,
		Terraform:  h.terraform,
		WorkingDir: h.repo.GetExecutePath(),
	}
	tfRunner, err := runner.NewLocalTerraform(h.client, h.terraform.Spec.Project)
	if err != nil {
		return "", errors.Wrapf(err, "new local terraform failed, terraform: %s, trace-id: %s", h.nn, h.traceId)
	}
	h.runner = tfRunner
	if _, err = h.runner.Init(req); err != nil {
		return "", errors.Wrapf(err, "new terraform failed, terraform: %s, trace-id: %s", h.nn, h.traceId)
	}

	/* init */
	_, err = h.runner.ExecInit(h.rootCtx, &runner.InitRequest{Upgrade: true, TfInstance: h.traceId})
	logs := h.runner.GetInitLog()
	if err != nil {
		return logs, errors.Wrapf(err, "init terraform failed, terraform: %s, trace-id: %s", h.nn, h.traceId)
	}
	blog.Info("init terraform runner success, terraform: %s, trace-id: %s", h.nn, h.traceId)

	h.terraform.Status.LastPlannedRevision = utils.FormatRevision(h.terraform.Spec.Repository.TargetRevision,
		h.repo.GetCommitId())
	h.terraform.Status.LastPlanAt = &metav1.Time{Time: time.Now()}

	/* plan */
	planReq := &runner.PlanRequest{
		Refresh:    false,
		Destroy:    false,
		TfInstance: h.traceId,
		Targets:    h.terraform.Spec.Targets,
	}
	planReply, err := h.runner.ExecPlan(h.rootCtx, planReq)
	logs = h.runner.GetPlanLog()
	if err != nil {
		return logs, errors.Wrapf(err, "plan terraform failed, terraform: %s, trace-id: %s", h.nn, h.traceId)
	}
	blog.Info("terraform plan message: %s, terraform: %s, trace-id: %s", planReply.Message, h.nn, h.traceId)

	return logs, nil
}

// saveTfPlanToSecret 保存结果到 secret
func (h *task) saveTfPlanToSecret() error {
	tfPlanBytes, err := os.ReadFile(h.runner.GetPlanOutFile())
	if err != nil {
		return errors.Wrapf(err, "read tf plan failed, tf: %s, trace-id: %s", h.nn, h.traceId)
	}
	// blog.Infof("tfPlanBytes: %d", len(tfPlanBytes)) //

	tfplanSecretExists := true
	tfplanSecret := new(corev1.Secret)
	secretName := fmt.Sprintf("tfplan-%s", h.terraform.Name)
	tfplanObjectKey := types.NamespacedName{Name: secretName, Namespace: h.terraform.Namespace}
	ctx, cancel := context.WithTimeout(h.rootCtx, 15*time.Second)
	defer cancel()

	if err = h.client.Get(ctx, tfplanObjectKey, tfplanSecret); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "unable to get the plan secret, tf: %s, trace-id: %s", h.nn, h.traceId)
		}
		tfplanSecretExists = false
		// blog.Info("secret not exists, name: %s", secretName)
	}
	if tfplanSecretExists {
		ctx, cancel := context.WithTimeout(h.rootCtx, 15*time.Second)
		defer cancel()
		blog.Infof("delete tfplan secret, name: %s, trace-id: %s", secretName, h.traceId)
		if err = h.client.Delete(ctx, tfplanSecret, &client.DeleteOptions{}); err != nil {
			return errors.Wrapf(err, "unable to delete the plan secret, tf: %s, trace-id: %s", h.nn, h.traceId)
		}
	}

	tfplan, err := utils.GzipEncode(tfPlanBytes)
	if err != nil {
		return errors.Wrapf(err, "unable to encode the plan revision, tf: %s, trace-id: %s", h.nn, h.traceId)
	}
	tfplanData := map[string][]byte{TFPlanName: tfplan}

	tfplanSecret2 := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: h.terraform.Namespace,
			Annotations: map[string]string{
				"encoding": "gzip",
				// 注解
				SavedPlanSecretAnnotation: h.repo.GetCommitId(), // 记录commit-id
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: tfv1.GroupVersion.String(),
					Kind:       tfv1.Kind,
					Name:       h.terraform.Name,
					UID:        h.terraform.UID,
				},
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: tfplanData,
	}
	if err = h.client.Create(h.rootCtx, tfplanSecret2, &client.CreateOptions{}); err != nil {
		return errors.Wrapf(err, "unable to create plan secret, tf: %s, trace-id: %s", h.nn, h.traceId)
	}
	blog.Info("save tfplan to secret success, secretName: %s, tf: %s, trace-id: %s", secretName, h.nn, h.traceId)

	return nil
}

// saveTfPlanToConfigMap 保存明文结果到configmap
func (h *task) saveTfPlanToConfigMap(raw string) error {
	tfplanCMExists := true
	tfplanCM := new(corev1.ConfigMap)
	configMapName := fmt.Sprintf("tfplan-%s", h.terraform.Name)
	tfplanObjectKey := types.NamespacedName{Name: configMapName, Namespace: h.terraform.Namespace}
	ctx, cancel := context.WithTimeout(h.rootCtx, 15*time.Second)
	defer cancel()

	if err := h.client.Get(ctx, tfplanObjectKey, tfplanCM); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "unable to get the plan configmap, tf: %s, trace-id: %s", h.nn, h.traceId)
		}
		tfplanCMExists = false
	}

	if tfplanCMExists {
		ctx, cancel := context.WithTimeout(h.rootCtx, 15*time.Second)
		defer cancel()
		blog.Infof("delete tfplan configmap, name: %s, trace-id: %s", configMapName, h.traceId)
		if err := h.client.Delete(ctx, tfplanCM); err != nil {
			return errors.Wrapf(err, "unable to delete the plan configmap, tf: %s, trace-id: %s", h.nn, h.traceId)
		}
	}

	tfplanData := map[string]string{TFPlanName: raw}
	tfplanCM2 := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: h.terraform.Namespace,
			Annotations: map[string]string{
				SavedPlanSecretAnnotation: h.repo.GetCommitId(),
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: tfv1.GroupVersion.String(),
					Kind:       tfv1.Kind,
					Name:       h.terraform.Name,
					UID:        h.terraform.UID,
				},
			},
		},
		Data: tfplanData,
	}

	if err := h.client.Create(h.rootCtx, &tfplanCM2); err != nil {
		return errors.Wrapf(err, "unable to create plan configmap, tf: %s, trace-id: %s", h.nn, h.traceId)
	}
	blog.Info("save tfplan to configmap success, configMapName: %s, tf: %s, trace-id: %s", configMapName, h.nn,
		h.traceId)

	return nil
}

// SaveTfPlan 保存plan结果
func (h *task) SaveTfPlan() error {
	raw, err := h.runner.ExecShowPlanFileRaw(h.rootCtx)
	if err != nil {
		return err
	}
	blog.Info("get tfplan success, tf: %s, trace-id: %s", h.nn, h.traceId)

	if err = h.saveTfPlanToSecret(); err != nil {
		return err
	}

	if err = h.saveTfPlanToConfigMap(raw); err != nil {
		return err
	}

	return nil
}

// Apply tf cr
func (h *task) Apply() (string, error) {
	h.terraform.Status.LastAppliedRevision = h.repo.GetCommitId()
	h.terraform.Status.LastAppliedAt = &metav1.Time{Time: time.Now()}
	h.terraform.Status.LastAttemptedRevision = utils.FormatRevision(h.terraform.Spec.Repository.TargetRevision,
		h.repo.GetCommitId())

	/* apply */
	applyReply, err := h.runner.ExecApply(h.rootCtx, &runner.ApplyRequest{TfInstance: h.traceId})
	logs := h.runner.GetApplyLog()
	if err != nil {
		return logs, errors.Wrapf(err, "apply terraform failed, terraform: %s, trace-id: %s", h.nn, h.traceId)
	}
	blog.Infof("terraform apply message: %s, terraform: %s, trace-id: %s", applyReply.Message, h.nn, h.traceId)

	return logs, nil
}

// SaveApplyOutputToConfigMap 保存apply输出结果
func (h *task) SaveApplyOutputToConfigMap(raw string) error {
	tfplanCMExists := true
	tfplanCM := new(corev1.ConfigMap)
	// note: 取短的commit-id或长的commit-id
	configMapName := fmt.Sprintf("tfapply-%s-%s", h.terraform.Name, h.lastCommitId[:8])
	tfplanObjectKey := types.NamespacedName{Name: configMapName, Namespace: h.terraform.Namespace}
	ctx, cancel := context.WithTimeout(h.rootCtx, 15*time.Second)
	defer cancel()

	if err := h.client.Get(ctx, tfplanObjectKey, tfplanCM); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "unable to get the plan configmap, tf: %s, trace-id: %s", h.nn, h.traceId)
		}
		tfplanCMExists = false
	}

	if tfplanCMExists {
		ctx, cancel := context.WithTimeout(h.rootCtx, 15*time.Second)
		defer cancel()
		blog.Infof("delete tfapply configmap, name: %s, trace-id: %s", configMapName, h.traceId)
		if err := h.client.Delete(ctx, tfplanCM); err != nil {
			return errors.Wrapf(err, "unable to delete the plan configmap, tf: %s, trace-id: %s", h.nn, h.traceId)
		}
	}

	tfplanData := map[string]string{TFApplyName: raw}
	tfplanCM2 := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: h.terraform.Namespace,
			Annotations: map[string]string{
				SavedApplySecretAnnotation: h.repo.GetCommitId(),
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: tfv1.GroupVersion.String(),
					Kind:       tfv1.Kind,
					Name:       h.terraform.Name,
					UID:        h.terraform.UID,
				},
			},
		},
		Data: tfplanData,
	}

	if err := h.client.Create(h.rootCtx, tfplanCM2); err != nil {
		return errors.Wrapf(err, "unable to create plan configmap, tf: %s, trace-id: %s", h.nn, h.traceId)
	}

	blog.Info("save tf apply to configmap success, tf: %s, trace-id: %s", h.nn, h.traceId)

	return nil
}

// Clean repository
func (h *task) Clean() {
	// note: 调试过程中，先注释掉
	// if err := h.repo.Clean(); err != nil {
	//	blog.Errorf("core handler clean repository failed, err: %s", err.Error())
	// }
}

// Stop 停止执行
func (h *task) Stop() {
	h.cancel()
}

// GetLastCommitId 获取分支最后一次commit-id
func (h *task) GetLastCommitId() (string, error) {
	if len(h.lastCommitId) != 0 {
		return h.lastCommitId, nil
	}

	commitId, err := h.repo.GetLastCommitId()
	if err != nil { // 获取最新的commit-id失败
		return "", err
	}
	h.lastCommitId = commitId

	return h.lastCommitId, nil
}

// getRepoToken get repo token
func (h *task) getRepoToken() error {
	secrets := new(corev1.SecretList)

	if err := h.client.List(h.rootCtx, secrets, client.InNamespace(h.terraform.Namespace)); err != nil {
		return errors.Wrapf(err, "list repo secrets failed, tf: %s", h.nn.String())
	}

	for _, item := range secrets.Items {
		url, ok := item.Data["url"]
		if !ok || string(url) != h.terraform.Spec.Repository.Repo { // url不存在 或者 是不等于repo url
			continue
		}
		if token, ok := item.Data["password"]; ok {
			h.repoToken = string(token)
			break
		}
	}

	if len(h.repoToken) == 0 {
		return errors.Errorf("not found '%s' repo password, tf: %s", h.terraform.Spec.Repository.Repo, h.nn.String())
	}

	return nil
}

//// Delete terraform
//// 1. new terraform
//// 1. destroy terraform
// func (h *task) Delete() error {
//	req := &runner.DestroyRequest{
//		TfInstance: h.traceId,
//		Targets:    h.terraform.Spec.Targets,
//	}
//	if _, err := h.tfRunner.Destroy(h.rootCtx, req); err != nil {
//		return errors.Wrapf(err, "destroy terraform failed, terraform: %s", h.nn)
//	}
//	blog.Info("destroy terraform success, terraform: %s", h.nn)
//
//	return nil
// }
