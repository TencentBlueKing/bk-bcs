package controller

import (
	"bk-bcs/bcs-mesos/pkg/controller/agentschedinfo"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, agentschedinfo.Add)
}
