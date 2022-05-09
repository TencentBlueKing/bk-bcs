package storage

import (
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage/tkex/gamedeployment/v1alpha1"
	gsv1alpha1 "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage/tkex/gamestatefulset/v1alpha1"
	schedtype "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/types"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"time"
)

type Namespace struct {
	CommonDataHeader
	Data *corev1.Namespace
}

type Deployment struct {
	CommonDataHeader
	Data *appv1.Deployment
}

type DaemonSet struct {
	CommonDataHeader
	Data *appv1.DaemonSet
}

type StatefulSet struct {
	CommonDataHeader
	Data *appv1.StatefulSet
}

type GameDeployment struct {
	CommonDataHeader
	Data *gdv1alpha1.GameDeployment
}

type GameStatefulSet struct {
	CommonDataHeader
	Data *gsv1alpha1.GameStatefulSet
}

type MesosApplication struct {
	CommonDataHeader
	Data *Application
}

type MesosDeployment struct {
	CommonDataHeader
	Data *schedtype.Deployment
}

type MesosNamespace string

type K8sNode struct {
	CommonDataHeader
	Data *corev1.Node
}

type Application struct {
	ID              string
	Name            string
	Metadata        commtypes.ObjectMeta
	DefineInstances uint64
	Instance        uint64
	RunningInstance uint64
	BuildedInstance int64
	RunAs           string
	ClusterId       string
	Status          string
	LastStatus      string
	CreateTime      time.Time
	UpdateTime      time.Time
	Mode            string
	LastUpdateTime  time.Time
	ReportTime      time.Time

	//we should replace the next three BcsXXX, using ObjectMeta.Labels directly
	BcsAppID    string
	BcsSetID    string
	BcsModuleID string

	Message string
	Pods    []*commtypes.BcsPodIndex
}
