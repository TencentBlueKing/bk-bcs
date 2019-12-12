package mesos

import (
	commtypes "bk-bcs/bcs-common/common/types"
)

type MesosInject interface {
	InjectApplicationContent(*commtypes.ReplicaController) (*commtypes.ReplicaController, error)
	InjectDeployContent(*commtypes.BcsDeployment) (*commtypes.BcsDeployment, error)
}
