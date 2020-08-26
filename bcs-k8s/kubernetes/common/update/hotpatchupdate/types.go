package hotpatchupdate

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// HotPatchUpdateStateKey records the state of hotpatch-update.
	// The value of annotation is HotPatchUpdateState.
	HotPatchUpdateStateKey string = "hotpatch-update-state"

	//hot-patch annotation
	PodHotpatchContainerKey = "io.kubernetes.hotpatch.container"
)

// HotPatchUpdateState records latest hotpatch-update state, including old statuses of containers.
type HotPatchUpdateState struct {
	// Revision is the updated revision hash.
	Revision string `json:"revision"`

	// UpdateTimestamp is the time when the hot-patch update happens.
	UpdateTimestamp metav1.Time `json:"updateTimestamp"`

	// LastContainerStatuses records the before-hot-patch-update container statuses. It is a map from ContainerName
	// to HotPatchUpdateContainerStatus
	LastContainerStatuses map[string]HotPatchUpdateContainerStatus `json:"lastContainerStatuses"`
}

// HotPatchUpdateContainerStatus records the statuses of the container that are mainly used
// to determine whether the HotPatchUpdate is completed.
type HotPatchUpdateContainerStatus struct {
	ImageID string `json:"imageID,omitempty"`
}
