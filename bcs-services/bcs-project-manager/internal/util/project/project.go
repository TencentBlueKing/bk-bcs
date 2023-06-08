package project

import (
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/cmdb"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

func PatchBusinessName(projects []*proto.Project) error {
	bizIDs := []int{}
	for _, project := range projects {
		// 历史遗留原因，以前迁移的一部分项目未开启容器服务，但是却设置了业务ID为0
		if project.Kind == "k8s" && project.BusinessID != "" && project.BusinessID != "0" {
			bizID, err := strconv.Atoi(project.BusinessID)
			if err != nil {
				return err
			}
			bizIDs = append(bizIDs, bizID)
		}
	}
	details, err := cmdb.BatchSearchBusinessByBizIDs(bizIDs)
	if err != nil {
		return err
	}
	businessMap := make(map[string]string)
	for _, biz := range details.Info {
		businessMap[strconv.Itoa(int(biz.BKBizID))] = biz.BKBizName
	}
	for _, project := range projects {
		if _, ok := businessMap[project.BusinessID]; !ok {
			project.BusinessName = ""
		}
		project.BusinessName = businessMap[project.BusinessID]
	}
	return nil
}
