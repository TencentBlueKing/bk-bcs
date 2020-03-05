/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package options

import (
	regd "bk-bcs/bcs-common/common/RegisterDiscover"
	"bk-bcs/bcs-common/common/blog"
	bcsssl "bk-bcs/bcs-common/common/ssl"
	"bk-bcs/bcs-common/common/static"
	"bk-bcs/bcs-common/common/types"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	restful "github.com/emicklei/go-restful"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/pflag"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type ServerType int

const (
	HTTP                          = "http"
	HTTPS                         = "https"
	ServerTypeSecure   ServerType = 1
	ServerTypeInsecure ServerType = 2

	TLSTypeClient = 1
	TLSTypeServer = 2
)

type TLSConfig struct {
	TLSType int

	CertFile string
	KeyFile  string
	CAFile   string
	// paasword of private key
	Password string
}

// ToServerConfigObj read current TLS Config files and return a Config object
func (c TLSConfig) ToConfigObj() (config *tls.Config, err error) {
	if c.CertFile == "" || c.CAFile == "" || c.KeyFile == "" {
		err = errors.New("missing argument, must provide all certfile/keyfile/cafile")
		return nil, err
	}

	switch c.TLSType {
	case TLSTypeClient:
		config, err = bcsssl.ClientTslConfVerity(c.CAFile, c.CertFile, c.KeyFile, c.Password)
	case TLSTypeServer:
		if c.Password == "" {
			c.Password = static.ServerCertPwd
		}
		blog.V(3).Infof("ca %s, cert %s, key %s, password %s", c.CAFile, c.CertFile, c.KeyFile, c.Password)
		config, err = bcsssl.ServerTslConfVerityClient(c.CAFile, c.CertFile, c.KeyFile, c.Password)
	}
	if err != nil || config == nil {
		return config, err
	}
	config.BuildNameToCertificate()
	return config, nil
}

type KubeDriverServerOptions struct {
	BindAddress      net.IP
	HostIP           string
	ZkServers        string
	SecurePort       uint
	InsecurePort     uint
	ExternalIp       string
	ExternalPort     uint
	KubeMasterUrl    string
	KubeClientTLS    TLSConfig
	ClusterClientTLS TLSConfig
	ClusterKeeperUrl string
	ServerTLS        TLSConfig
	RootWebContainer *restful.Container
	Environment      string

	// 181114 added by wesleylin, support custom report ip:port and clusterID
	CustomReportAddress string
	CustomReportPort    uint
	CustomClusterID     string
}

func NewKubeDriverServerOptions() *KubeDriverServerOptions {
	return &KubeDriverServerOptions{
		ClusterClientTLS: TLSConfig{TLSType: TLSTypeClient},
		KubeClientTLS:    TLSConfig{TLSType: TLSTypeClient},
		ServerTLS:        TLSConfig{TLSType: TLSTypeServer},
	}
}

// BindFlagSet binds a ServerOptions with a FlagSet, when it get parsed, ServerOptions's values will be set
// by input flags.
func (o *KubeDriverServerOptions) BindFlagSet(fs *pflag.FlagSet) {
	fs.StringVar(&o.Environment, "environment", "prod", "Environment, prod default, (prod, stag, develop). Set develop to avoid failure of fetching clusterID")
	fs.IPVar(&o.BindAddress, "address", net.ParseIP("127.0.0.1"), "The ip address for the serve on")
	fs.StringVar(&o.HostIP, "host-ip", "", "host ip which is used.")

	fs.UintVar(&o.InsecurePort, "insecure-port", 0, "The insecure port for the serve on, such as 30001.")
	fs.UintVar(&o.SecurePort, "secure-port", 0, "The secure port for the serve on, such as 30443.")

	fs.StringVar(&o.ExternalIp, "external-ip", "", "external IP address to listen on for this service.")
	fs.UintVar(&o.ExternalPort, "external-port", 0, "external port to listen on for this service")

	// k8s related
	fs.StringVar(&o.KubeMasterUrl, "kube-master-url", "", "The host of the Kubernetes ApiServer"+
		"to connect to in the format of scheme://address:port, e.g., http://localhost:8080")
	fs.StringVar(&o.KubeClientTLS.CAFile, "kube-ca-file", "", "kube-master trusted root certificates.")
	fs.StringVar(&o.KubeClientTLS.CertFile, "kube-cert-file", "", "kube-master requires TLS client certificate.")
	fs.StringVar(&o.KubeClientTLS.KeyFile, "kube-key-file", "", "kube-master requires TLS client certificate.")

	// zk
	fs.StringVar(&o.ZkServers, "zk-url", "", "zk url. eg http://127.0.0.1:2181,http://127.0.0.1:2181.")

	// cluster keeper cls
	fs.StringVar(&o.ClusterClientTLS.CAFile, "cluster-ca-file", "", "cluster keeper trusted root certificates.")
	fs.StringVar(&o.ClusterClientTLS.CertFile, "cluster-cert-file", "", "cluster keeper requires TLS certificate.")
	fs.StringVar(&o.ClusterClientTLS.KeyFile, "cluster-key-file", "", "cluster keeper requires TLS certificate.")
	fs.StringVar(&o.ClusterClientTLS.Password, "cluster-key-password", "", "cluster keeper requires TLS certificate.")

	// server tls
	fs.StringVar(&o.ServerTLS.CAFile, "server-ca-file", "", "server trusted root certificates.")
	fs.StringVar(&o.ServerTLS.CertFile, "server-cert-file", "", "server requires TLS certificate.")
	fs.StringVar(&o.ServerTLS.KeyFile, "server-key-file", "", "server requires TLS certificate.")
	fs.StringVar(&o.ServerTLS.Password, "server-key-password", "", "server tls private key password.")

	// custom
	fs.StringVar(&o.CustomReportAddress, "custom-report-address", "", "custom bind address to report to zk")
	fs.UintVar(&o.CustomReportPort, "custom-report-port", 0, "custom bind port to report to zk")
	fs.StringVar(&o.CustomClusterID, "custom-cluster-id", "", "custom clusterID")

}

func (o *KubeDriverServerOptions) SecureServerConfigured() bool {
	return o.SecurePort > 0
}

func (o *KubeDriverServerOptions) InsecureServerConfigured() bool {
	return o.InsecurePort > 0
}

// Validate checks given parameters is valid or not
func (o *KubeDriverServerOptions) Validate() error {
	if !(o.SecureServerConfigured() || o.InsecureServerConfigured()) {
		return errors.New("you must provide at least one of secure-port and insecure-port")
	}
	if o.KubeMasterUrl == "" {
		return errors.New("kube-master-url can not be empty")
	}

	_, err := url.Parse(o.KubeMasterUrl)
	if err != nil {
		return fmt.Errorf("not a valid URL address: %s", o.KubeMasterUrl)
	}

	if o.ZkServers == "" && o.Environment != "develop" {
		return errors.New("zk-url can not be empty")
	}
	return nil
}

func (o *KubeDriverServerOptions) NeedClientTLSConfig() bool {
	kubeURL, _ := url.Parse(o.KubeMasterUrl)
	return kubeURL.Scheme == HTTPS
}

func (o *KubeDriverServerOptions) NeedClusterTLSConfig() bool {
	clusterURL, _ := url.Parse(o.ClusterKeeperUrl)
	return clusterURL.Scheme == HTTPS
}

// MakeServerAddress returns server address like "127.0.0.1:8080" for http module to listen on
func (o *KubeDriverServerOptions) MakeServerAddress(serverType ServerType) string {
	var port uint
	if serverType == ServerTypeSecure {
		port = o.SecurePort
	} else {
		port = o.InsecurePort
	}
	return net.JoinHostPort(o.BindAddress.String(), strconv.FormatUint(uint64(port), 10))
}

func (o *KubeDriverServerOptions) GetClusterKeeperAddr() error {
	blog.Infof("start to get cluster keeper api addr.")
	disc := regd.NewRegDiscoverEx(o.ZkServers, time.Duration(5*time.Second))
	if err := disc.Start(); nil != err {
		return fmt.Errorf("start get cluster keeper  zk service failed. Error:%v", err)
	}

	watchPath := fmt.Sprintf("%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_CLUSTERKEEPER)
	eventChan, eventErr := disc.DiscoverService(watchPath)
	if nil != eventErr {
		return fmt.Errorf("start running discover service failed error:%v", eventErr)
	}
	defer func() {
		if err := disc.Stop(); nil != err {
			blog.Errorf("stop get cluster keeper  addr zk discover failed. reason: %v", err)
		}
	}()
	for {
		select {
		case data := <-eventChan:
			blog.Info("received one zk event which may contains cluster keeper  address.")
			if data.Err != nil {
				return fmt.Errorf("get cluster keeper  api failed. reason: %s", data.Err.Error())
			}
			if len(data.Server) == 0 {
				return errors.New("get 0 cluster keeper  api address.")
			}
			info := types.ServerInfo{}
			if err := json.Unmarshal([]byte(data.Server[0]), &info); nil != err {
				return fmt.Errorf("unmashal cluster keeper  server info failed. reason: %v", err)
			}
			if len(info.IP) == 0 || info.Port == 0 || len(info.Scheme) == 0 {
				return fmt.Errorf("get invalid cluster keeper  info: %s", data.Server[0])
			}
			clusterKeeperUrl := fmt.Sprintf("%s://%s:%d", info.Scheme, info.IP, info.Port)
			blog.V(3).Infof("get valid cluster keeper  url: %s", clusterKeeperUrl)
			o.ClusterKeeperUrl = clusterKeeperUrl
			return nil
			//case <-timeout:
			//	return "", errors.New("watch cluster keeper  api address timeout.")
		default:
			time.Sleep(time.Duration(1 * time.Second))
			blog.V(3).Info("try to get cluster keeper api address, waiting for a second.")
		}
	}

}
