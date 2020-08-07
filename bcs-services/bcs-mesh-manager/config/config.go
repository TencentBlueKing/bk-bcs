package config

type Config struct {
	IstioOperatorNs string
	IstioOperatorName string
	DockerHub string
	IstioOperatorCrFile string
	IstioOperatorCrdFile string
	//bcs api-gateway address
	ServerAddress string
	//api-gateway usertoken
	UserToken string
}