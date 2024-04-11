package handler

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/middleware/xbknodeman"
	pb "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/proto"
)

type BcsApiHandler struct {
	bkAppCode   string
	bkAppSecret string
	bkEnv       string
}

func NewBcsApiHandler(bkAppCode, bkAppSecret, bkEnv string) *BcsApiHandler {
	return &BcsApiHandler{
		bkAppCode:   bkAppCode,
		bkAppSecret: bkAppSecret,
		bkEnv:       bkEnv,
	}
}

func (b *BcsApiHandler) InstallJob(ctx context.Context, request *pb.InstallJobRequest,
	response *pb.InstallJobResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	installReq := &xbknodeman.InstallJobRequest{}
	for _, host := range request.Hosts {
		installReq.Hosts = append(installReq.Hosts, &xbknodeman.InstallHost{
			BkCloudId: host.BkCloudId,
			BkBizId:   host.BkBizId,
			BkHostID:  host.BkHostId,
			OsType:    host.OsType,
			InnerIp:   host.InnerIp,
			OuterIp:   host.OuterIp,
			LoginIp:   host.LoginIp,
			Account:   host.Account,
			Port:      host.Port,
			AuthType:  host.AuthType,
			Password:  host.Password,
			ApId:      host.ApId,
			Key:       host.Key,
		})
	}
	installReq.JobType = request.JobType
	resp, err := nodeManCli.InstallJob(ctx, installReq)
	if err != nil {
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}
	response.Code = common.CodeSuccess
	response.Message = resp.Message
	response.Data = &pb.InstallJobData{
		JobId: resp.Data.JobId,
	}
	return nil
}

func (b *BcsApiHandler) CreateCloud(ctx context.Context, request *pb.CloudCreateRequest,
	response *pb.CloudCreateResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	resp, err := nodeManCli.CreateCloud(ctx, &xbknodeman.CreateCloudRequest{
		BkCloudName: request.BkCloudName,
		Isp:         request.Isp,
		ApID:        int64(request.ApId),
	})
	if err != nil {
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	response.Code = common.CodeSuccess
	response.Message = resp.Message
	response.Data = int32(resp.Data.BkCloudID)
	return nil
}

func (b *BcsApiHandler) UpdateCloud(ctx context.Context, request *pb.CloudUpdateRequest,
	response *pb.CloudUpdateResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	resp, err := nodeManCli.UpdateCloud(ctx, &xbknodeman.UpdateCloudRequest{
		BkCloudID:   int64(request.BkCloudId),
		BkCloudName: request.BkCloudName,
		Isp:         request.Isp,
		ApID:        int64(request.ApId),
	})
	if err != nil {
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	response.Code = common.CodeSuccess
	response.Message = resp.Message
	return nil
}

func (b *BcsApiHandler) ListCloud(ctx context.Context, request *pb.CloudListRequest,
	response *pb.CloudListResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	resp, err := nodeManCli.ListCloud(ctx, &xbknodeman.ListCloudRequest{})
	if err != nil {
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	response.Code = common.CodeSuccess
	response.Message = resp.Message
	for _, cloud := range resp.Data {
		response.Data = append(response.Data, &pb.Cloud{
			BkCloudId:   int32(cloud.BkCloudId),
			BkCloudName: cloud.BkCloudName,
			Isp:         cloud.Isp,
			ApId:        int32(cloud.ApId),
		})
	}
	return nil
}

func (b *BcsApiHandler) DeleteCloud(ctx context.Context, request *pb.CloudDeleteRequest,
	response *pb.CloudDeleteResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	resp, err := nodeManCli.DeleteCloud(ctx, &xbknodeman.DeleteCloudRequest{
		BkCloudID: int64(request.BkCloudId),
	})
	if err != nil {
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	response.Code = common.CodeSuccess
	response.Message = resp.Message
	return nil
}

func (b *BcsApiHandler) ListHost(ctx context.Context, request *pb.ListHostRequest,
	response *pb.ListHostResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	listReq := &xbknodeman.ListHostRequest{
		Page:     int64(request.Page),
		PageSize: int64(request.Pagesize),
	}

	for _, cond := range request.Conditions {
		listReq.Conditions = append(listReq.Conditions, xbknodeman.Condition{
			Key:   cond.Key,
			Value: cond.Value,
		})
	}

	resp, err := nodeManCli.ListHosts(ctx, listReq)
	if err != nil {
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	err = common.JsonConvert(resp.Data, &response.Data)
	if err != nil {
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}
	response.Code = common.CodeSuccess
	response.Message = resp.Message
	return nil
}

func (b *BcsApiHandler) ListProxyHost(ctx context.Context, request *pb.ListProxyHostRequest,
	response *pb.ListProxyHostResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	listReq := &xbknodeman.GetProxyHostRequest{
		BkCloudId: request.BkCloudId,
	}

	resp, err := nodeManCli.GetProxyHost(ctx, listReq)
	if err != nil {
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	err = common.JsonConvert(resp.Data, &response.Data)
	if err != nil {
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}
	response.Code = common.CodeSuccess
	response.Message = resp.Message
	return nil
}

func (b *BcsApiHandler) GetJobDetail(ctx context.Context, request *pb.GetJobDetailRequest,
	response *pb.GetJobDetailResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	blog.Errorf("req: %s", common.JsonMarshal(request))
	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	jobDetailReq := &xbknodeman.GetJobDetailRequest{
		JobID:    request.JobId,
		Page:     int64(request.Page),
		PageSize: int64(request.Pagesize),
	}

	for _, cond := range request.Conditions {
		jobDetailReq.Conditions = append(jobDetailReq.Conditions, xbknodeman.Condition{
			Key:   cond.Key,
			Value: cond.Value,
		})
	}
	jobDetail, err := nodeManCli.GetJobDetails(ctx, jobDetailReq)
	if err != nil {
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	err = common.JsonConvert(jobDetail.Data, &response.Data)
	if err != nil {
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}
	response.Code = common.CodeSuccess
	response.Message = jobDetail.Message
	return nil
}

func (b *BcsApiHandler) newBkNodeManCli(userName string) *xbknodeman.Client {
	return xbknodeman.NewClient(0, b.bkEnv, b.bkAppCode, b.bkAppSecret, "", userName)
}
