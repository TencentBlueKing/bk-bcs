package v1

import (
	"context"
	"fmt"
	clientmeshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/meshmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"
	"google.golang.org/grpc/credentials"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type meshManager struct {
	clientOption types.ClientOptions
	meshManagerClient meshmanager.MeshManagerClient
	conn *grpc.ClientConn
}

//NewMeshManager
func NewMeshManager(options types.ClientOptions)clientmeshmanager.MeshManager{
	m := &meshManager{
		clientOption: options,
	}
	return m
}

func (m *meshManager) dialGrpc()error{
	var err error
	addr := strings.TrimLeft(m.clientOption.BcsApiAddress, "http://")
	addr = strings.TrimLeft(m.clientOption.BcsApiAddress, "https://")
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"authorization": fmt.Sprintf("Bearer %s", m.clientOption.BcsToken),
	}
	md := metadata.New(header)
	m.conn, err = grpc.Dial(
		addr,
		grpc.WithDefaultCallOptions(grpc.Header(&md)),
		grpc.WithTransportCredentials(credentials.NewTLS(m.clientOption.ClientSSL)),
		/*grpc.WithPerRPCCredentials(utils.GrpcTokenAuth{
			Token: m.clientOption.
		,
		}),*/
	)
	if err != nil {
		return err
	}
	m.meshManagerClient = meshmanager.NewMeshManagerClient(m.conn)
	return nil
}

func (m *meshManager) closeGrpc(){
	m.conn.Close()
	m.conn = nil
	m.meshManagerClient = nil
}

func (m *meshManager) CreateMeshCluster(req *meshmanager.CreateMeshClusterReq)(*meshmanager.CreateMeshClusterResp,error){
	err := m.dialGrpc()
	if err!=nil {
		return nil, err
	}
	defer m.closeGrpc()

	return m.meshManagerClient.CreateMeshCluster(context.TODO(), req)
}

func (m *meshManager) DeleteMeshCluster(req *meshmanager.DeleteMeshClusterReq)(*meshmanager.DeleteMeshClusterResp,error){
	err := m.dialGrpc()
	if err!=nil {
		return nil, err
	}
	defer m.closeGrpc()

	return m.meshManagerClient.DeleteMeshCluster(context.TODO(), req)
}

func (m *meshManager) ListMeshCluster(req *meshmanager.ListMeshClusterReq)(*meshmanager.ListMeshClusterResp,error){
	err := m.dialGrpc()
	if err!=nil {
		return nil, err
	}
	defer m.closeGrpc()

	return m.meshManagerClient.ListMeshCluster(context.TODO(), req)
}