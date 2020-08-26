package core

import (
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/update/inplaceupdate"
	v1 "k8s.io/api/core/v1"
)

type Control interface {
	// common
	IsInitializing() bool
	SetRevisionTemplate(revisionSpec map[string]interface{}, template map[string]interface{})
	ApplyRevisionPatch(patched []byte) (*tkexv1alpha1.GameDeployment, error)

	// scale
	IsReadyToScale() bool
	NewVersionedPods(currentCS, updateCS *tkexv1alpha1.GameDeployment,
		currentRevision, updateRevision string,
		expectedCreations, expectedCurrentCreations int,
		availableIDs []string,
	) ([]*v1.Pod, error)

	// update
	IsPodUpdatePaused(pod *v1.Pod) bool
	IsPodUpdateReady(pod *v1.Pod, minReadySeconds int32) bool
	GetPodsSortFunc(pods []*v1.Pod, waitUpdateIndexes []int) func(i, j int) bool
	GetUpdateOptions() *inplaceupdate.UpdateOptions

	// validation
	ValidateGameDeploymentUpdate(oldCS, newCS *tkexv1alpha1.GameDeployment) error
}

func New(gd *tkexv1alpha1.GameDeployment) Control {
	return &commonControl{GameDeployment: gd}
}
