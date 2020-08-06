package k8s

import "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"

type ClusterLogController struct {
	clusterInfo *bcsapi.ClusterCredential
	tick        int64
}

func NewClusterLogController(cc *bcsapi.ClusterCredential) (*ClusterLogController, error) {
	ctlr := &ClusterLogController{
		clusterInfo: cc,
		tick:        0,
	}
	return ctlr, nil
}

func (c *ClusterLogController) Start() {
}

func (c *ClusterLogController) Stop() {
}

func (c *ClusterLogController) SetTick(tick int64) {
	c.tick = tick
}

func (c *ClusterLogController) GetTick() int64 {
	return c.tick
}
