package app

import (
	"crypto/tls"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
)

func Run(op *options.LogManagerOption) error {
	conf := &config.Config{}
	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}
	err := setManagerConfig(op, conf)
	if err != nil {
		blog.Errorf("Parse Manager config error %s", err.Error())
		os.Exit(1)
	}
	manager, err := k8s.NewManager(conf)
	if err != nil {
		blog.Errorf("NewManager error %s", err.Error())
		os.Exit(1)
	}

	manager.Start()
	blog.Info("Log Manager started...")
	return nil
}

func setManagerConfig(op *options.LogManagerOption, conf *config.Config) error {
	conf.CollectionConfigs = op.CollectionConfigs
	tlsconf, err := tls.LoadX509KeyPair(op.ClientCertFile, op.ClientKeyFile)
	if err != nil {
		return err
	}
	conf.BcsApiConfig.Host = op.BcsApiHost
	conf.BcsApiConfig.AuthToken = op.AuthToken
	conf.BcsApiConfig.Gateway = op.Gateway
	conf.BcsApiConfig.TLSConfig = tlsconf
	return nil
}
