package handler

import (
	"context"

	mesh "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"

	"k8s.io/klog"
)

type Mesh struct{}

// CreateMeshCluster is a single request handler called via client.Call or the generated client code
func (e *Mesh) CreateMeshCluster(ctx context.Context, req *mesh.CreateMeshClusterReq) (*mesh.CreateMeshClusterResp, error) {
	klog.Infof("Received MeshManager.CreateMeshCluster request %s", req.String())


}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (e *Mesh) Stream(ctx context.Context, req *mesh.StreamingRequest, stream mesh.Mesh_StreamStream) error {
	log.Infof("Received Mesh.Stream request with count: %d", req.Count)

	for i := 0; i < int(req.Count); i++ {
		log.Infof("Responding: %d", i)
		if err := stream.Send(&mesh.StreamingResponse{
			Count: int64(i),
		}); err != nil {
			return err
		}
	}

	return nil
}

// PingPong is a bidirectional stream handler called via client.Stream or the generated client code
func (e *Mesh) PingPong(ctx context.Context, stream mesh.Mesh_PingPongStream) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		log.Infof("Got ping %v", req.Stroke)
		if err := stream.Send(&mesh.Pong{Stroke: req.Stroke}); err != nil {
			return err
		}
	}
}
