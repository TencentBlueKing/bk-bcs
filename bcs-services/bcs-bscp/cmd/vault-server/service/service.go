package service

import (
	"context"
	"fmt"
	"net/http"

	pbvs "bscp.io/pkg/protocol/vault-server"
	"bscp.io/pkg/serviced"
	"github.com/pkg/errors"
)

// Service do all the data service's work
type Service struct {
	gateway *gateway
}

// NewService create a service instance.
func NewService(sd serviced.Discover) (*Service, error) {

	state, ok := sd.(serviced.State)
	if !ok {
		return nil, errors.New("discover convert state failed")
	}
	gateway, err := newGateway(state)
	if err != nil {
		return nil, fmt.Errorf("new gateway failed, err: %v", err)
	}

	s := &Service{

		gateway: gateway,
	}

	return s, nil
}

// Handler return service's handler.
func (s *Service) Handler() (http.Handler, error) {
	if s.gateway == nil {
		return nil, errors.New("gateway is nil")
	}

	return s.gateway.handler(), nil
}

// Ping .
func (s *Service) Ping(ctx context.Context, in *pbvs.PingMsg) (*pbvs.PingMsg, error) {
	return nil, nil
}
