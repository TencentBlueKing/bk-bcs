package service

import (
	"context"
	"strconv"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbhook "bscp.io/pkg/protocol/core/hook"
	pbds "bscp.io/pkg/protocol/data-service"
)

// CreateToken create a token
func (s *Service) CreateToken(ctx context.Context, req *pbcs.CreateTokenReq) (*pbcs.CreateTokenResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateTokenResp)

	bizID, err := strconv.Atoi(grpcKit.SpaceID)
	if err != nil {
		return nil, err
	}
	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Token, Action: meta.Create}, BizID: uint32(bizID)}
	err = s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	masterKey := cc.AuthServer().Token.MasterKey
	encryption_algorithm := cc.AuthServer().Token.Ea

	r := &pbds.CreateTokenReq{
		Memo:  req.Memo,
		BizId: uint32(bizID),
		Rule:  req.Rule,
	}
	rp, err := s.client.DS.CreateToken(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.CreateHookResp{
		Id: rp.Id,
	}
	return resp, nil
}
