package service

import (
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	hr "bscp.io/pkg/protocol/core/hook-release"
	pbhr "bscp.io/pkg/protocol/core/hook-release"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
	"context"
	"time"
)

func (s *Service) CreateHookRelease(ctx context.Context,
	req *pbds.CreateHookReleaseReq) (*pbds.CreateResp, error) {

	kt := kit.FromGrpcContext(ctx)
	// TODO 获取是否已经存在
	//if _, err := s.dao.HookRelease().GetByName(kt, req.Attachment.BizId, req.Spec.Name); err == nil {
	//	return nil, fmt.Errorf("templateSpace name %s already exists", req.Spec.Name)
	//}

	spec, err := req.Spec.HookReleaseSpec()
	if err != nil {
		logs.Errorf("get HookReleaseSpec spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	now := time.Now()
	hookRelease := &table.HookRelease{
		Spec:       spec,
		Attachment: req.Attachment.HookReleaseAttachment(),
		Revision: &table.Revision{
			Creator:   kt.User,
			Reviser:   kt.User,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	id, err := s.dao.HookRelease().Create(kt, hookRelease)
	if err != nil {
		logs.Errorf("create HookRelease failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

func (s *Service) ListHookReleases(ctx context.Context,
	req *pbds.ListHookReleasesReq) (*pbds.ListHookReleasesResp, error) {

	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	details, count, err := s.dao.HookRelease().List(kt, req.BizId, req.HookId, opt)
	if err != nil {
		logs.Errorf("list HookRelease failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	hookRelease, err := pbhr.PbHookReleaseSpaces(details)
	if err != nil {
		logs.Errorf("get pb hookRelease failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListHookReleasesResp{
		Count:   uint32(count),
		Details: hookRelease,
	}
	return resp, nil
}

func (s *Service) GetHookReleaseByID(ctx context.Context,
	req *pbds.GetHookReleaseByIdReq) (*hr.HookRelease, error) {

	kt := kit.FromGrpcContext(ctx)

	hookRelease, err := s.dao.HookRelease().Get(kt, req.GetBizId(), req.GetHookId())
	if err != nil {
		logs.Errorf("get app by id failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp, _ := hr.PbHookRelease(hookRelease)
	return resp, nil
}

func (s *Service) DeleteHookRelease(ctx context.Context,
	req *pbds.DeleteHookReleaseReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	HookRelease := &table.HookRelease{
		ID: req.Id,
		Attachment: &table.HookReleaseAttachment{
			BizID:  req.BizId,
			HookID: req.HookId,
		},
	}

	if err := s.dao.HookRelease().Delete(kt, HookRelease); err != nil {
		logs.Errorf("delete HookRelease failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

func (s *Service) PublishHookRelease(ctx context.Context, req *pbds.PublishHookReleaseReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)
	now := time.Now()
	HookRelease := &table.HookRelease{
		ID: req.Id,
		Attachment: &table.HookReleaseAttachment{
			BizID:  req.BizId,
			HookID: req.HookId,
		},
		Spec: &table.HookReleaseSpec{},
		Revision: &table.Revision{
			Reviser:   kt.User,
			UpdatedAt: now,
		},
	}

	if err := s.dao.HookRelease().Publish(kt, HookRelease); err != nil {
		logs.Errorf("delete HookRelease failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil

}
