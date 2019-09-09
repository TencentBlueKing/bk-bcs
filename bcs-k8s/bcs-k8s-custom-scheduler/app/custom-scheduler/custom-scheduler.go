package custom_scheduler

import (
	"bk-bcs/bcs-common/common/http/httpserver"
	"bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/config"
	"bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/pkg/actions"

	"fmt"
)

type CustomScheduler struct {
	config   *config.IpschedulerConfig
	httpServ *httpserver.HttpServer
}

func NewCustomScheduler(conf *config.IpschedulerConfig) *CustomScheduler {
	customSched := &CustomScheduler{
		config:   conf,
		httpServ: httpserver.NewHttpServer(conf.Port, conf.Address, conf.Sock),
	}

	if conf.ServCert.IsSSL {
		customSched.httpServ.SetSsl(conf.ServCert.CAFile, conf.ServCert.CertFile, conf.ServCert.KeyFile, conf.ServCert.CertPasswd)
	}

	customSched.httpServ.SetInsecureServer(conf.InsecureAddress, conf.InsecurePort)

	return customSched
}

func (p *CustomScheduler) Start() error {

	p.httpServ.RegisterWebServer("", nil, actions.GetApiAction())
	router := p.httpServ.GetRouter()
	webContainer := p.httpServ.GetWebContainer()
	router.Handle("/{sub_path:.*}", webContainer)
	if err := p.httpServ.ListenAndServeMux(p.config.VerifyClientTLS); err != nil {
		return fmt.Errorf("http ListenAndServe error %s", err.Error())
	}

	return nil
}
