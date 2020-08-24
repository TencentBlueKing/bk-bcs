package config

type Config struct {
	//IstioOperator Docker Hub
	DockerHub string
	//Istio Operator Charts
	IstioOperatorCharts string
	//IstioOperator cr
	IstioOperatorCr string
	//bcs api-gateway address
	ServerAddress string
	//api-gateway usertoken
	UserToken string
}