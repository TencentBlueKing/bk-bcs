package ipscheduler

import (
	"bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/config"
	"bk-bcs/bcs-services/bcs-netservice/pkg/netservice"
	"bk-bcs/bcs-services/bcs-netservice/pkg/netservice/types"

	"fmt"
	"strings"
)

//BcsConfig config item for ipscheduler
type BcsConfig struct {
	ZkHost   string         `json:"zkHost"`
	TLS      *types.SSLInfo `json:"tls,omitempty"`
	Interval int            `json:"interval,omitempty"`
}

func createNetSvcClient() (netservice.Client, error) {
	conf := newConf()

	var client netservice.Client
	var clientErr error
	if conf.TLS == nil {
		client, clientErr = netservice.NewClient()
	} else {
		client, clientErr = netservice.NewTLSClient(conf.TLS.CACert, conf.TLS.Key, conf.TLS.PubKey, conf.TLS.Passwd)
	}
	if clientErr != nil {
		return nil, clientErr
	}
	//client get bcs-netservice info
	hosts := strings.Split(conf.ZkHost, ";")
	if err := client.GetNetService(hosts); err != nil {
		return nil, fmt.Errorf("get netservice failed, %s", err.Error())
	}
	return client, nil
}

func newConf() BcsConfig {
	conf := BcsConfig{
		ZkHost: config.ZkHosts,
		TLS: &types.SSLInfo{
			CACert: config.ClientCert.CAFile,
			Key:    config.ClientCert.KeyFile,
			PubKey: config.ClientCert.CertFile,
			Passwd: config.ClientCert.CertPasswd,
		},
	}

	return conf
}
