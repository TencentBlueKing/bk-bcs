package subscriber

import (
	"context"

	log "github.com/micro/go-micro/v2/logger"

	mesh "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"
)

type Mesh struct{}

func (e *Mesh) Handle(ctx context.Context, msg *mesh.Message) error {
	log.Info("Handler Received message: ", msg.Say)
	return nil
}

func Handler(ctx context.Context, msg *mesh.Message) error {
	log.Info("Function Received message: ", msg.Say)
	return nil
}
