package controllers

import (
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	tkexclientset "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/client/clientset/versioned"
	gamedeploylister "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/client/listers/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
)

// GameDeploymentStatusUpdaterInterface is an interface used to update the GameDeploymentStatus associated with a GameDeployment.
// For any use other than testing, clients should create an instance using NewRealGameDeploymentStatusUpdater.
type GameDeploymentStatusUpdaterInterface interface {
	// UpdateGameDeploymentStatus sets the set's Status to status. Implementations are required to retry on conflicts,
	// but fail on other errors. If the returned error is nil set's Status has been successfully set to status.
	UpdateGameDeploymentStatus(deploy *tkexv1alpha1.GameDeployment, newStatus *tkexv1alpha1.GameDeploymentStatus, pods []*v1.Pod) error
}

// NewRealGameDeploymentStatusUpdater returns a GameDeploymentStatusUpdaterInterface that updates the Status of a GameDeployment,
// using the supplied client and setLister.
func NewRealGameDeploymentStatusUpdater(
	tkexClient tkexclientset.Interface,
	setLister gamedeploylister.GameDeploymentLister) GameDeploymentStatusUpdaterInterface {
	return &realGameDeploymentStatusUpdater{tkexClient, setLister}
}

type realGameDeploymentStatusUpdater struct {
	tkexClient tkexclientset.Interface
	setLister  gamedeploylister.GameDeploymentLister
}

func (r *realGameDeploymentStatusUpdater) UpdateGameDeploymentStatus(
	deploy *tkexv1alpha1.GameDeployment,
	newStatus *tkexv1alpha1.GameDeploymentStatus,
	pods []*v1.Pod) error {
	r.calculateStatus(deploy, newStatus, pods)
	if !r.inconsistentStatus(deploy, newStatus) {
		return nil
	}

	klog.Infof("To update GameDeployment status for  %s/%s, replicas=%d ready=%d available=%d updated=%d updatedReady=%d, revisions update=%s",
		deploy.Namespace, deploy.Name, newStatus.Replicas, newStatus.ReadyReplicas, newStatus.AvailableReplicas, newStatus.UpdatedReplicas, newStatus.UpdatedReadyReplicas, newStatus.UpdateRevision)
	return r.updateStatus(deploy, newStatus)
}

func (r *realGameDeploymentStatusUpdater) updateStatus(deploy *tkexv1alpha1.GameDeployment, newStatus *tkexv1alpha1.GameDeploymentStatus) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		deploy.Status = *newStatus
		_, updateErr := r.tkexClient.TkexV1alpha1().GameDeployments(deploy.Namespace).UpdateStatus(deploy)

		return updateErr
	})
}

func (r *realGameDeploymentStatusUpdater) inconsistentStatus(deploy *tkexv1alpha1.GameDeployment, newStatus *tkexv1alpha1.GameDeploymentStatus) bool {
	oldStatus := deploy.Status
	return newStatus.ObservedGeneration > oldStatus.ObservedGeneration ||
		newStatus.Replicas != oldStatus.Replicas ||
		newStatus.ReadyReplicas != oldStatus.ReadyReplicas ||
		newStatus.AvailableReplicas != oldStatus.AvailableReplicas ||
		newStatus.UpdatedReadyReplicas != oldStatus.UpdatedReadyReplicas ||
		newStatus.UpdatedReplicas != oldStatus.UpdatedReplicas ||
		newStatus.UpdateRevision != oldStatus.UpdateRevision ||
		newStatus.LabelSelector != oldStatus.LabelSelector
}

func (r *realGameDeploymentStatusUpdater) calculateStatus(deploy *tkexv1alpha1.GameDeployment, newStatus *tkexv1alpha1.GameDeploymentStatus, pods []*v1.Pod) {
	for _, pod := range pods {
		newStatus.Replicas++
		if util.IsRunningAndReady(pod) {
			newStatus.ReadyReplicas++
		}
		if util.IsRunningAndAvailable(pod, deploy.Spec.MinReadySeconds) {
			newStatus.AvailableReplicas++
		}
		if util.GetPodRevision(pod) == newStatus.UpdateRevision {
			newStatus.UpdatedReplicas++
		}
		if util.IsRunningAndReady(pod) && util.GetPodRevision(pod) == newStatus.UpdateRevision {
			newStatus.UpdatedReadyReplicas++
		}
	}
}
