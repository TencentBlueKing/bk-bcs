package app

import (
	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/app/custom-scheduler"
	"bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/config"
	"bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/options"
	"os"
)

//Run the ipscheduler
func Run(op *options.ServerOption) {

	conf := parseConfig(op)

	customSched := custom_scheduler.NewCustomScheduler(conf)
	//start customSched, and http service
	err := customSched.Start()
	if err != nil {
		blog.Errorf("start processor error %s, and exit", err.Error())
		os.Exit(1)
	}

	//pid
	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}

	return
}

func parseConfig(op *options.ServerOption) *config.IpschedulerConfig {
	ipschedulerConfig := config.NewIpschedulerConfig()

	ipschedulerConfig.Address = op.Address
	ipschedulerConfig.Port = op.Port
	ipschedulerConfig.InsecureAddress = op.InsecureAddress
	ipschedulerConfig.InsecurePort = op.InsecurePort
	ipschedulerConfig.ZkHosts = op.BCSZk
	ipschedulerConfig.VerifyClientTLS = op.VerifyClientTLS

	config.ZkHosts = op.BCSZk
	config.Cluster = op.Cluster
	config.Kubeconfig = op.Kubeconfig
	config.KubeMaster = op.KubeMaster
	config.UpdatePeriod = op.UpdatePeriod

	//server cert directory
	if op.CertConfig.ServerCertFile != "" && op.CertConfig.ServerKeyFile != "" {
		ipschedulerConfig.ServCert.CertFile = op.CertConfig.ServerCertFile
		ipschedulerConfig.ServCert.KeyFile = op.CertConfig.ServerKeyFile
		ipschedulerConfig.ServCert.CAFile = op.CertConfig.CAFile
		ipschedulerConfig.ServCert.IsSSL = true
	}

	//client cert directory
	if op.CertConfig.ClientCertFile != "" && op.CertConfig.ClientKeyFile != "" {
		ipschedulerConfig.ClientCert.CertFile = op.CertConfig.ClientCertFile
		ipschedulerConfig.ClientCert.KeyFile = op.CertConfig.ClientKeyFile
		ipschedulerConfig.ClientCert.CAFile = op.CertConfig.CAFile
		ipschedulerConfig.ClientCert.IsSSL = true
	}

	config.ClientCert = ipschedulerConfig.ClientCert

	return ipschedulerConfig
}
