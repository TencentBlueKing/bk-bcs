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

package tfhandler

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/internal/logctx"
	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/terraformextensions/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/repository"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/utils"
)

// TerraformHandler 定义 Terraform 处理的接口内容
type TerraformHandler interface {
	Destroy(ctx context.Context, tf *tfv1.Terraform) error
	Plan(ctx context.Context, tf *tfv1.Terraform, commitID string) (bool, error)
	Apply(ctx context.Context, tf *tfv1.Terraform, commitID string, historyID int) error

	GetPlanResult(ctx context.Context, tf *tfv1.Terraform) (*TerraformPlanOrApply, error)
	GetLastApply(ctx context.Context, tf *tfv1.Terraform) (*TerraformPlanOrApply, error)
}

// NewTerraformHandler 创建 TerraformHandler 实例
func NewTerraformHandler(repoHandler repository.Handler, k8sClient *kubernetes.Clientset) TerraformHandler {
	return &terraformHandler{
		repoHandler: repoHandler,
		k8sClient:   k8sClient,
	}
}

type terraformHandler struct {
	repoHandler repository.Handler
	k8sClient   *kubernetes.Clientset
}

// Destroy 销毁对应 Terraform 的内容
func (h *terraformHandler) Destroy(ctx context.Context, tf *tfv1.Terraform) error {
	// 未曾 apply 成功过的 cr 不需要销毁资源
	if tf.Status.LastAppliedRevision == "" {
		logctx.Infof(ctx, "terraform '%s/%s' not have last_applied_revision, no need destroy resources",
			tf.Namespace, tf.Name)
		return nil
	}
	workerDir := option.GetRepoStoragePath(tf.Name, string(tf.UID))
	if err := h.createWorkerDir(workerDir); err != nil {
		return err
	}
	repoPath, err := h.repoHandler.CheckoutCommit(ctx, &tf.Spec.Repository, tf.Status.LastAppliedRevision, workerDir)
	if err != nil {
		return errors.Wrapf(err, "checkout commit to '%s' failed", workerDir)
	}
	tfExec, err := newTerraformExec(tf, repoPath)
	if err != nil {
		return errors.Wrapf(err, "create terraform client failed")
	}
	if err = tfExec.ExecInit(ctx); err != nil {
		return err
	}
	applyResult, err := tfExec.ExecApply(ctx, true)
	if err != nil {
		return err
	}
	logctx.Infof(ctx, "destroy result: %v", applyResult)
	return nil
}

// Plan 获取 Terraform 的 plan 结果, 返回值为 True 表示存在变化，为 false 表示无变化
func (h *terraformHandler) Plan(ctx context.Context, tf *tfv1.Terraform, commitID string) (bool, error) {
	workerDir := option.GetRepoStoragePath(tf.Name, string(tf.UID))
	if err := h.createWorkerDir(workerDir); err != nil {
		return false, err
	}
	repoPath, err := h.repoHandler.CheckoutCommit(ctx, &tf.Spec.Repository, commitID, workerDir)
	if err != nil {
		return false, errors.Wrapf(err, "checkout commit to '%s' failed", workerDir)
	}

	logctx.Infof(ctx, "terraform planning")
	tfExec, err := newTerraformExec(tf, repoPath)
	if err != nil {
		return false, errors.Wrapf(err, "create terraform client failed")
	}
	if err = tfExec.ExecInit(ctx); err != nil {
		return false, err
	}
	planResult, hasChanged, err := tfExec.ExecPlan(ctx)
	if err != nil {
		return false, err
	}
	if !hasChanged {
		return hasChanged, nil
	}
	if err = h.saveTerraformPlanOrApplyToKubernetes(ctx, typePlan, &TerraformPlanOrApply{
		Namespace: tf.Namespace,
		CommitID:  commitID,
		Result:    planResult,
		OwnerName: tf.Name,
		OwnerUID:  tf.UID,
	}); err != nil {
		return false, errors.Wrapf(err, "save terraform plan secret failed")
	}
	return hasChanged, nil
}

// Apply 执行 Apply 的动作，检查是否已经存在 Worker 目录，存在则不进行 Init 动作
func (h *terraformHandler) Apply(ctx context.Context, tf *tfv1.Terraform, commitID string, historyID int) error {
	workerDir := option.GetRepoStoragePath(tf.Name, string(tf.UID))
	defer func() {
		if err := h.deleteWorkerDir(workerDir); err != nil {
			logctx.Warnf(ctx, "delete worker directory '%s' failed after apply", workerDir)
		}
	}()
	logctx.Infof(ctx, "terraform applying")
	var tfExec *TerraformExec
	if !h.isExistWorkerDir(ctx, workerDir) {
		logctx.Infof(ctx, "worker dir '%s' not exist, will auto-create", workerDir)
		if err := h.createWorkerDir(workerDir); err != nil {
			return err
		}
		repoPath, err := h.repoHandler.CheckoutCommit(ctx, &tf.Spec.Repository, commitID, workerDir)
		if err != nil {
			return errors.Wrapf(err, "checkout commit to '%s' failed", workerDir)
		}
		tfExec, err = newTerraformExec(tf, repoPath)
		if err != nil {
			return errors.Wrapf(err, "create terraform client failed")
		}
		if err = tfExec.ExecInit(ctx); err != nil {
			return err
		}
	} else {
		repoPath := workerDir + "/" + utils.ParseGitRepoName(tf.Spec.Repository.Repo)
		var err error
		tfExec, err = newTerraformExec(tf, repoPath)
		if err != nil {
			return errors.Wrapf(err, "create terraform client failed")
		}
	}
	applyResult, err := tfExec.ExecApply(ctx, false)
	if err != nil {
		return err
	}
	if err = h.saveTerraformPlanOrApplyToKubernetes(ctx, typeApply, &TerraformPlanOrApply{
		Namespace:      tf.Namespace,
		CommitID:       commitID,
		Result:         applyResult,
		OwnerName:      tf.Name,
		OwnerUID:       tf.UID,
		ApplyHistoryID: historyID,
	}); err != nil {
		return errors.Wrapf(err, "save terraform apply secret failed")
	}
	return nil
}

func (h *terraformHandler) createWorkerDir(workerDir string) error {
	if err := os.MkdirAll(workerDir, 0655); err != nil {
		return errors.Wrapf(err, "create worker dir '%s' failed", workerDir)
	}
	return nil
}

func (h *terraformHandler) deleteWorkerDir(workerDir string) error {
	if err := os.RemoveAll(workerDir); err != nil {
		return errors.Wrapf(err, "delete worker dir '%s' failed", workerDir)
	}
	return nil
}

func (h *terraformHandler) isExistWorkerDir(ctx context.Context, workerDir string) bool {
	if _, err := os.Stat(workerDir); err != nil {
		if !os.IsNotExist(err) {
			logctx.Warnf(ctx, "os.stat worker dir '%s' failed: %s", workerDir, err.Error())
		}
		return false
	}
	return true
}

// TerraformPlanOrApply 定义对应 Plan 或 Apply 的结果
type TerraformPlanOrApply struct {
	Namespace    string    `json:"namespace"`
	CommitID     string    `json:"commitID"`
	Result       string    `json:"result"`
	CreationTime time.Time `json:"creationTime"`

	ApplyHistoryID int `json:"applyHistoryID"`

	OwnerName string    `json:"-"`
	OwnerUID  types.UID `json:"-"`
}

type terraformType string

const (
	terraformPlanPrefix  = "tfplan"
	terraformApplyPrefix = "tfapply"

	typePlan  terraformType = "plan"
	typeApply terraformType = "apply"
)

func terraformApplySecretName(tfName string, historyID int) string {
	return utils.TruncateString(terraformApplyPrefix+"-"+tfName+"-"+strconv.Itoa(historyID), 63, 16)
}

func terraformPlanSecretName(tfName string) string {
	return utils.TruncateString(terraformPlanPrefix+"-"+tfName, 63, 16)
}

// GetPlanResult 或者 Plan 的结果
func (h *terraformHandler) GetPlanResult(ctx context.Context, tf *tfv1.Terraform) (*TerraformPlanOrApply, error) {
	secretName := terraformPlanSecretName(tf.Name)
	secret, err := h.k8sClient.CoreV1().Secrets(tf.Namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, errors.Wrapf(err, "get terraform plan secret '%s' failed", secretName)
		}
		return nil, nil
	}
	tfObj, err := transSecretToTerraformObj(secret)
	if err != nil {
		return nil, errors.Wrapf(err, "trans secret to terraform plan failed")
	}
	return tfObj, nil
}

// GetLastApply 获取 Apply 的结果
func (h *terraformHandler) GetLastApply(ctx context.Context, tf *tfv1.Terraform) (*TerraformPlanOrApply, error) {
	secretName := terraformApplySecretName(tf.Name, tf.Status.History.ID)
	secret, err := h.k8sClient.CoreV1().Secrets(tf.Namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, errors.Wrapf(err, "get terraform apply secret '%s' failed", secretName)
		}
		return nil, nil
	}
	tfObj, err := transSecretToTerraformObj(secret)
	if err != nil {
		return nil, errors.Wrapf(err, "trans secret to terraform plan failed")
	}
	return tfObj, nil
}

func (h *terraformHandler) saveTerraformPlanOrApplyToKubernetes(ctx context.Context, tfType terraformType,
	tfObj *TerraformPlanOrApply) error {
	transSecret, err := transTerraformObjToSecret(tfObj, tfType)
	if err != nil {
		return errors.Wrapf(err, "trans terraform obj to secret failed")
	}
	secretName := transSecret.ObjectMeta.Name
	secret, err := h.k8sClient.CoreV1().Secrets(tfObj.Namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return errors.Wrapf(err, "get secret '%s' failed", secretName)
		}
		_, err = h.k8sClient.CoreV1().Secrets(tfObj.Namespace).Create(ctx, transSecret, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrapf(err, "create secret '%s' failed", secretName)
		}
		logctx.Infof(ctx, "create secret '%s' success", secretName)
		return nil
	}
	secret.Data = transSecret.Data
	if _, err = h.k8sClient.CoreV1().Secrets(tfObj.Namespace).Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
		return errors.Wrapf(err, "update secret '%s' failed", secretName)
	}
	logctx.Infof(ctx, "update secret '%s' success", secretName)
	return nil
}

func transTerraformObjToSecret(tfObj *TerraformPlanOrApply, tfType terraformType) (*corev1.Secret, error) {
	gzipResult, err := utils.GzipEncode([]byte(tfObj.Result))
	if err != nil {
		return nil, errors.Wrapf(err, "gzip terraform plan result failed")
	}
	var secretName string
	switch tfType {
	case typePlan:
		secretName = terraformPlanSecretName(tfObj.OwnerName)
	case typeApply:
		secretName = terraformApplySecretName(tfObj.OwnerName, tfObj.ApplyHistoryID)
	}
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: tfObj.Namespace,
			Annotations: map[string]string{
				"encoding": "gzip",
			},
			Labels: map[string]string{
				"terraform.bkbcs.tencent.com/terraform-name": tfObj.OwnerName,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: tfv1.SchemeGroupVersion.String(),
					Kind:       tfv1.TerraformKind,
					Name:       tfObj.OwnerName,
					UID:        tfObj.OwnerUID,
				},
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"commit": []byte(tfObj.CommitID),
			"result": gzipResult,
		},
	}
	if tfType == typeApply {
		secret.Data["id"] = []byte(strconv.Itoa(tfObj.ApplyHistoryID))
	}
	return secret, nil
}

func transSecretToTerraformObj(secret *corev1.Secret) (*TerraformPlanOrApply, error) {
	resultBS := secret.Data["result"]
	bs, err := utils.GzipDecode(resultBS)
	if err != nil {
		return nil, errors.Wrapf(err, "gzip decode terraform plan failed")
	}
	return &TerraformPlanOrApply{
		Namespace:    secret.Namespace,
		CommitID:     string(secret.Data["commit"]),
		Result:       string(bs),
		CreationTime: secret.CreationTimestamp.Time,
	}, nil
}
