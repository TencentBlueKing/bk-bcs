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
 *
 */

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/argoproj/argo-cd/v2/util/db"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/internal/logctx"
	tfproto "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/proto"
	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/terraformextensions/v1"
	tfclient "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/client/clientset/versioned"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/repository"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/tfhandler"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/worker/podrunner"
)

// TerraformWorker defines the worker of terraform handle
type TerraformWorker struct {
	index int
	op    *option.ControllerOption
	conn  *grpc.ClientConn

	repoHandler repository.Handler
	tfHandler   tfhandler.TerraformHandler
	tfk8sClient *tfclient.Clientset
	k8sClient   *kubernetes.Clientset
	argoDB      db.ArgoDB
}

// parsePodIndex will parse the worker name, and get the indx from it. Such as: "bcs-terraform-worker-1"
func (w *TerraformWorker) parsePodIndex() error {
	podIndex := os.Getenv(podrunner.EnvPodIndex)
	tmp := strings.Split(podIndex, "-")
	index := tmp[len(tmp)-1]
	var err error
	w.index, err = strconv.Atoi(index)
	if err != nil {
		return errors.Wrapf(err, "parse POD_INDEX '%s' failed", podIndex)
	}
	return nil
}

// Init will init some plugins and init the connection to grpc server
func (w *TerraformWorker) Init(ctx context.Context) error {
	w.op = option.GlobalOption()
	if err := w.parsePodIndex(); err != nil {
		return err
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return errors.Wrapf(err, "get in-cluster config failed")
	}
	w.k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "create in-cluster client failed")
	}
	w.tfk8sClient, err = tfclient.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "create terraform in-cluster client failed")
	}

	argoDB, _, err := store.NewArgoDB(ctx, w.op.ArgoAdminNamespace)
	if err != nil {
		return errors.Wrapf(err, "create argo db failed")
	}
	w.argoDB = argoDB
	w.repoHandler = repository.NewRepositoryHandler(w.argoDB)
	w.tfHandler = tfhandler.NewTerraformHandler(w.repoHandler, w.k8sClient)

	w.conn, err = grpc.Dial(w.op.ControllerGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.Wrapf(err, "grpc dial '%s' failed", w.op.ControllerGRPCAddress)
	}
	return nil
}

// Start will auto-poll terraform message from grpc-server and handle it.
func (w *TerraformWorker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logctx.Warnf(ctx, "worker will be done with context done")
			return
		default:
			terraform, err := w.poll(ctx)
			if err != nil {
				logctx.Errorf(ctx, "poll terraform object from grpc failed: %s", err.Error())
				time.Sleep(time.Second)
				continue
			}
			if terraform == nil {
				time.Sleep(5 * time.Second)
				continue
			}
			ctx = context.WithValue(ctx, logctx.TraceKey, uuid.New().String())
			ctx = context.WithValue(ctx, logctx.ObjectKey, terraform.Namespace+"/"+terraform.Name)
			if err = w.handler(ctx, terraform); err != nil {
				terraform.Status.OperationStatus = tfv1.OperationStatus{
					Message: err.Error(),
					Phase:   tfv1.PhaseError,
				}
				logctx.Errorf(ctx, "handle terraform failed: %s", err.Error())
			} else {
				logctx.Infof(ctx, "handle terraform success")
				terraform.Status.OperationStatus = tfv1.OperationStatus{Phase: tfv1.PhaseSucceeded}
			}
			terraform.Status.OperationStatus.FinishAt = &metav1.Time{Time: time.Now()}
			if err = w.updateStatusOperation(ctx, terraform); err != nil {
				logctx.Errorf(ctx, "update status operation failed: %s", err.Error())
			}
		}
	}
}

// poll the terraform message from grpc server with idx
func (w *TerraformWorker) poll(ctx context.Context) (*tfv1.Terraform, error) {
	if w.conn.GetState() != connectivity.Ready && w.conn.GetState() != connectivity.Idle {
		logctx.Warnf(ctx, "grpc connection not ready but '%s', need reconnect", w.conn.GetState().String())
		var err error
		_ = w.conn.Close()
		w.conn, err = grpc.Dial(w.op.ControllerGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, errors.Wrapf(err, "grpc reconnect failed")
		}
	}
	rpcClient := tfproto.NewQueueClient(w.conn)
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	resp, err := rpcClient.Poll(ctx, &tfproto.PollRequest{
		Index: int32(w.index),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "poll from grpc server failed")
	}
	if resp.Data == nil {
		logctx.Warnf(ctx, "not get message from queue")
		return nil, nil
	}
	tf := new(tfv1.Terraform)
	if err = json.Unmarshal(resp.Data, tf); err != nil {
		return nil, errors.Wrapf(err, "unmarshal failed: %s", string(resp.Data))
	}
	return tf, nil
}

// handler 处理 terraform 对象
// 1. 获取对应仓库分支的最新提交，执行 Plan 的动作
// 2. 确认是否需要执行 Apply 动作：
//   - 确认是否有 TerraformOperationSync Annotation，有则为手动同步，需要 Apply
//   - 确认是否为自动同步，是则需要 Apply
//
// 3. 执行 Apply 动作
func (w *TerraformWorker) handler(ctx context.Context, terraform *tfv1.Terraform) error {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 600*time.Second)
	defer cancel()
	startedAt := time.Now()

	tfRepo := terraform.Spec.Repository
	lastCommitID, err := w.repoHandler.GetLastCommitId(ctx, &tfRepo)
	if err != nil {
		return errors.Wrapf(err, "get last commit id failed")
	}
	logctx.Infof(ctx, "got last commit for '%s, %s': %s", tfRepo.Repo, tfRepo.TargetRevision, lastCommitID)
	if err = w.execPlan(ctx, terraform, lastCommitID); err != nil {
		return errors.Wrapf(err, "execute terraform plan failed")
	}

	// 需要清理资源，并自动设置同步策略为 manual
	if cleanValue, ok := terraform.Annotations[tfv1.TerraformOperationClean]; ok {
		return w.execClean(ctx, terraform, cleanValue)
	}

	needApply := false
	if syncCommitID, ok := terraform.Annotations[tfv1.TerraformOperationSync]; ok {
		if err = w.patchDeleteTerraformOperation(ctx, terraform, tfv1.TerraformOperationSync); err != nil {
			return errors.Wrapf(err, "patch failed")
		}
		logctx.Infof(ctx, "need sync to '%s' by manual sync annotation", syncCommitID)
		if terraform.Status.SyncStatus == tfv1.SyncedStatus {
			logctx.Infof(ctx, "terraform is synced, no need sync again")
			return nil
		}
		if syncCommitID != lastCommitID {
			logctx.Warnf(ctx, "manual sync to '%s' not same with last-commit '%s', no need apply",
				syncCommitID, lastCommitID)
		} else {
			needApply = true
		}
	}
	if terraform.Status.SyncStatus == tfv1.SyncedStatus {
		return nil
	}
	if terraform.Spec.SyncPolicy == tfv1.AutoSyncPolicy {
		needApply = true
	}
	if !needApply {
		return nil
	}
	if err = w.execApply(ctx, terraform, lastCommitID); err != nil {
		return err
	}
	terraform.Status.History = tfv1.ApplyHistory{
		ID:         terraform.Status.History.ID + 1,
		StartedAt:  &metav1.Time{Time: startedAt},
		FinishedAt: &metav1.Time{Time: time.Now()},
		Revision:   lastCommitID,
	}
	return nil
}

// execApply 执行 Apply 动作
func (w *TerraformWorker) execApply(ctx context.Context, tf *tfv1.Terraform, lastCommitID string) error {
	defer func() {
		if err := w.updateStatusApply(ctx, tf); err != nil {
			logctx.Warnf(ctx, "update status apply failed: %s", err.Error())
		}
	}()
	tf.Status.LastAppliedRevision = lastCommitID
	tf.Status.LastAppliedAt = &metav1.Time{Time: time.Now()}
	tf.Status.LastApplyError = ""
	if err := w.tfHandler.Apply(ctx, tf, lastCommitID, tf.Status.History.ID+1); err != nil {
		tf.Status.LastApplyError = err.Error()
		return errors.Wrapf(err, "terraform apply failed")
	}
	tf.Status.SyncStatus = tfv1.SyncedStatus
	return nil
}

func (w *TerraformWorker) execClean(ctx context.Context, tf *tfv1.Terraform, cleanValue string) error {
	logctx.Infof(ctx, "need clean managed-resources of terraform with clean annotation value: %s",
		cleanValue)
	if err := w.patchDeleteTerraformOperation(ctx, tf, tfv1.TerraformOperationClean); err != nil {
		return errors.Wrapf(err, "patch failed")
	}
	if err := w.tfHandler.Destroy(ctx, tf); err != nil {
		return errors.Wrapf(err, "terraform destroy failed")
	}
	return nil
}

// execPlan 通过 Plan 检测管理的资源是否发生了变化，即使 CommitID 一致仍然可能因为线上资源发生变化（删除）而 存在差异
// 若存在差异，则将同步状态修改为 OutOfSync
func (w *TerraformWorker) execPlan(ctx context.Context, terraform *tfv1.Terraform, lastCommitID string) error {
	defer func() {
		if err := w.updateStatusPlan(ctx, terraform); err != nil {
			logctx.Warnf(ctx, "patch status plan failed: %s", err.Error())
		}
	}()
	if lastCommitID != terraform.Status.LastAppliedRevision {
		terraform.Status.SyncStatus = tfv1.OutOfSyncStatus
	}
	terraform.Status.LastPlannedRevision = lastCommitID
	terraform.Status.LastPlannedAt = &metav1.Time{Time: time.Now()}
	terraform.Status.LastPlanError = ""
	changed, err := w.tfHandler.Plan(ctx, terraform, lastCommitID)
	if err != nil {
		terraform.Status.LastPlanError = err.Error()
		return errors.Wrapf(err, "terraform plan failed")
	}
	if changed {
		terraform.Status.SyncStatus = tfv1.OutOfSyncStatus
	} else {
		terraform.Status.SyncStatus = tfv1.SyncedStatus
	}
	return nil
}

func (w *TerraformWorker) updateStatusPlan(ctx context.Context, terraform *tfv1.Terraform) error {
	newTf, err := w.tfk8sClient.TerraformextensionsV1().Terraforms(terraform.Namespace).
		Get(ctx, terraform.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "get terraform failed")
	}
	newTf.Status.LastPlannedRevision = terraform.Status.LastPlannedRevision
	newTf.Status.LastPlannedAt = terraform.Status.LastPlannedAt
	newTf.Status.LastPlanError = terraform.Status.LastPlanError
	newTf.Status.SyncStatus = terraform.Status.SyncStatus
	_, err = w.tfk8sClient.TerraformextensionsV1().Terraforms(terraform.Namespace).
		UpdateStatus(ctx, newTf, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrapf(err, "update status plan failed")
	}
	logctx.Infof(ctx, "update status operation plan success")
	return nil
}

func (w *TerraformWorker) updateStatusApply(ctx context.Context, terraform *tfv1.Terraform) error {
	newTf, err := w.tfk8sClient.TerraformextensionsV1().Terraforms(terraform.Namespace).
		Get(ctx, terraform.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "get terraform failed")
	}
	newTf.Status.LastAppliedRevision = terraform.Status.LastAppliedRevision
	newTf.Status.LastAppliedAt = terraform.Status.LastAppliedAt
	newTf.Status.LastApplyError = terraform.Status.LastApplyError
	newTf.Status.SyncStatus = terraform.Status.SyncStatus
	_, err = w.tfk8sClient.TerraformextensionsV1().Terraforms(terraform.Namespace).
		UpdateStatus(ctx, newTf, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrapf(err, "update status apply failed")
	}
	logctx.Infof(ctx, "update status plan success")
	return nil
}

func (w *TerraformWorker) updateStatusOperation(ctx context.Context, terraform *tfv1.Terraform) error {
	newTf, err := w.tfk8sClient.TerraformextensionsV1().Terraforms(terraform.Namespace).
		Get(ctx, terraform.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "get terraform failed")
	}
	newTf.Status.OperationStatus = terraform.Status.OperationStatus
	newTf.Status.History = terraform.Status.History
	_, err = w.tfk8sClient.TerraformextensionsV1().Terraforms(terraform.Namespace).
		UpdateStatus(ctx, newTf, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrapf(err, "update status operation failed")
	}
	logctx.Infof(ctx, "update status operation success")
	return nil
}

func (w *TerraformWorker) patchDeleteTerraformOperation(ctx context.Context,
	tf *tfv1.Terraform, operation string) error {
	patches := []byte(fmt.Sprintf(`[{"op":"remove","path":"/metadata/annotations/%s"}]`, operation))
	_, err := w.tfk8sClient.TerraformextensionsV1().Terraforms(tf.Namespace).
		Patch(ctx, tf.Name, types.JSONPatchType, patches, metav1.PatchOptions{})
	if err != nil {
		return errors.Wrapf(err, "patch delete operation '%s' failed", operation)
	}
	logctx.Infof(ctx, "patch delete operation annotation sucess")
	return nil
}
