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
	//address
	Address string
	//port, grpc port, http port +1
	Port int
	//metrics port
	MetricsPort string
	//etcd cert file
	EtcdCertFile string
	//etcd key file
	EtcdKeyFile string
	//etcd ca file
	EtcdCaFile string
	//server ca file
	ServerCaFile string
	//server key file
	ServerKeyFile string
	//server cert file
	ServerCertFile string
}