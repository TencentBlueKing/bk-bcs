package config

import (
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

//MeshManagerOptions new meshmanager options to parse command-line parameters
type MeshManagerOptions struct {
	conf.FileConfig
	conf.MetricConfig
	conf.ServiceConfig
	conf.CertConfig

	DockerHub           string `json:"istio-docker-hub" value:"" usage:"istio-operator docker hub"`
	IstioOperatorCharts string `json:"istiooperator-charts" value:"" usage:"istio-operator charts"`
	IstioConfiguration  string `json:"istio-configuration" value:"" usage:"istio configuration"`
	ServerAddress       string `json:"apigateway-addr" value:"" usage:"bcs apigateway address"`
	UserToken           string `json:"user-token" value:"" usage:"bcs apigateway usertoken to control k8s cluster"`
	Kubeconfig          string `json:"kubeconfig" value:"" usage:"kube-apiserver kubeconfig"`
	EtcdCaFile          string `json:"etcd-cafile" value:"" usage:"SSL Certificate Authority file used to secure etcd communication"`
	EtcdCertFile        string `json:"etcd-certfile" value:"" usage:"SSL certification file used to secure etcd communication"`
	EtcdKeyFile         string `json:"etcd-keyfile" value:"" usage:"SSL key file used to secure etcd communication"`
	EtcdServers         string `json:"etcd-servers" value:"" usage:"List of etcd servers to connect with (scheme://ip:port), comma separated"`
}

//ParseConfig parse command-line parameters to mesh-manager config struct
func ParseConfig() Config {
	op := &MeshManagerOptions{}
	conf.Parse(op)
	conf := Config{}
	conf.Address = op.Address
	conf.Port = op.Port
	conf.MetricsPort = strconv.Itoa(int(op.MetricPort))
	conf.DockerHub = op.DockerHub
	conf.IstioOperatorCharts = op.IstioOperatorCharts
	conf.ServerAddress = op.ServerAddress
	conf.UserToken = op.UserToken
	conf.EtcdCaFile = op.EtcdCaFile
	conf.EtcdCertFile = op.EtcdCertFile
	conf.EtcdKeyFile = op.EtcdKeyFile
	conf.EtcdServers = op.EtcdServers
	conf.Kubeconfig = op.Kubeconfig
	conf.IstioConfiguration = op.IstioConfiguration
	//server cert directory
	if op.CertConfig.ServerCertFile != "" && op.CertConfig.ServerKeyFile != "" {
		conf.ServerCertFile = op.CertConfig.ServerCertFile
		conf.ServerKeyFile = op.CertConfig.ServerKeyFile
		conf.ServerCaFile = op.CertConfig.CAFile
	}
	return conf
}

// ValidateConfig check nessacessry
func ValidateConfig() error {
	//! for config item safety
	return nil
}
