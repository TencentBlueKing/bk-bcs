package app

import (
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bkdata "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/bkdataapi"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
)

func Run(op *options.LogManagerOption) error {
	conf := &config.ManagerConfig{}
	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}
	controller := bkdata.NewBKDataController(op.KubeConfig)
	err := controller.Start()
	if err != nil {
		blog.Errorf("BKDataApi controller start failed: %s", err.Error())
		os.Exit(1)
	}
	blog.Info("BKDataApiConfig controller started")
	err = setManagerConfig(op, conf)
	if err != nil {
		blog.Errorf("Parse Manager config error %s", err.Error())
		os.Exit(1)
	}
	manager := k8s.NewManager(conf)

	manager.Start()
	blog.Info("Log Manager started...")
	return nil
}

func setManagerConfig(op *options.LogManagerOption, conf *config.ManagerConfig) error {
	conf.CollectionConfigs = op.CollectionConfigs
	conf.BcsApiConfig.Host = op.BcsAPIHost
	conf.BcsApiConfig.AuthToken = op.AuthToken
	conf.BcsApiConfig.Gateway = op.Gateway
	conf.CAFile = op.ClientCertFile
	conf.SystemDataID = op.SystemDataID
	conf.BkAppCode = op.BkAppCode
	conf.BkUsername = op.BkUsername
	conf.BkAppSecret = op.BkAppSecret
	conf.BkBizID = op.BkBizID
	conf.KubeConfig = op.KubeConfig
	return nil
}
