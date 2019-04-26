## CommonConfig 及各组件启动参数统一方案

1. 都需要有`--file`加载启动参数的方式，文件格式为json
2. 共有的参数保持一致，引用common/conf中的结构
3. 不启用多余的，不需要的参数
4. 各模块自定义的参数格式，小写+下划线
5. 需要export成为flag的，tag格式为：json:"flag_name" short:"short_hand_of_flag" value:"default_value" usage:"flag_usage"，其中json,value,usage为必须字段
6. flag支持多级结构体嵌套
7. flag支持IntSlice和StringSlice，默认值用逗号(,)分割，如 value:"a,b,c"
8. 参数解析优先级：命令行 > 配置文件 > 默认值

参考例子（storage启动参数配置）:
```golang
type StorageOptions struct {
    // 引用所需参数
    conf.FileConfig
    conf.ServiceConfig
    conf.MetricConfig
    conf.ZkConfig
    conf.ServerOnlyCertConfig

    // storage自定义参数
    DBConfig     string `json:"database_config_file" value:"storage-database.conf" usage:"Config file for database."`
}

func main() {
    op := &StorageOptions{}

    // 解析参数
    conf.Parse(op)
}
```

目前共有的参数包括
```golang
// Config file, if set it will cover all the flag value it contains
type FileConfig struct {
    ConfigFile string `json:"file" short:"f" value:"" usage:"json file with configuration"`
}

// Service bind
type ServiceConfig struct {
    Address string `json:"address" short:"a" value:"127.0.0.1" usage:"IP address to listen on for this service"`
    Port    uint   `json:"port" short:"p" value:"8080" usage:"Port to listen on for this service"`
}

// Local info
type LocalConfig struct {
    LocalIP string `json:"local_ip" value:"127.0.0.1" usage:"IP address of this host"`
}

// Metric info
type MetricConfig struct {
    MetricPort uint `json:"metric_port" value:"8081" usage:"Port to listen on for metric"`
}

// Register discover
type ZkConfig struct {
    BCSZk string `json:"bcs_zookeeper" value:"127.0.0.1:2181" usage:"Zookeeper server for registering and discovering"`
}

// Server and client TLS config, can not be import with ClientCertOnlyConfig or ServerCertOnlyConfig
type CertConfig struct {
    ServerCertDir  string `json:"server_cert_dir" value:"" usage:"Directory of server certificate. If set, it will looking for bcs-inner-server.crt/bcs-inner-server.key/bcs-inner-ca.crt and set up an HTTPS server"`
    ClientCertDir  string `json:"client_cert_dir" value:"" usage:"Directory of client certificate. If set, it will looking for bcs-inner-client.crt/bcs-inner-client.key/bcs-inner-ca.crt"`
    CAFile         string `json:"ca_file" value:"" usage:"CA file. If server_cert_file/server_key_file/ca_file are all set, it will set up an HTTPS server required and verified client cert"`
    ServerCertFile string `json:"server_cert_file" value:"" usage:"Server public key file(*.crt). If both server_cert_file and server_key_file are set, it will set up an HTTPS server"`
    ServerKeyFile  string `json:"server_key_file" value:"" usage:"Server private key file(*.key). If both server_cert_file and server_key_file are set, it will set up an HTTPS server"`
    ClientCertFile string `json:"client_cert_file" value:"" usage:"Client public key file(*.crt)."`
    ClientKeyFile  string `json:"client_key_file" value:"" usage:"Client private key file(*.key)."`
}

// Client TLS config, can not be import with CertConfig or ServerCertOnlyConfig
type ClientOnlyCertConfig struct {
    ClientCertDir  string `json:"client_cert_dir" value:"" usage:"Directory of client certificate. If set, it will looking for bcs-inner-client.crt/bcs-inner-client.key/bcs-inner-ca.crt"`
    CAFile         string `json:"ca_file" value:"" usage:"CA file. If server_cert_file/server_key_file/ca_file are all set, it will set up an HTTPS server required and verified client cert"`
    ClientCertFile string `json:"client_cert_file" value:"" usage:"Client public key file(*.crt)."`
    ClientKeyFile  string `json:"client_key_file" value:"" usage:"Client private key file(*.key)."`
}

// Server TLS config, can not be import with ClientCertOnlyConfig or CertConfig
type ServerOnlyCertConfig struct {
    ServerCertDir  string `json:"server_cert_dir" value:"" usage:"Directory of server certificate. If set, it will looking for bcs-inner-server.crt/bcs-inner-server.key/bcs-inner-ca.crt and set up an HTTPS server"`
    CAFile         string `json:"ca_file" value:"" usage:"CA file. If server_cert_file/server_key_file/ca_file are all set, it will set up an HTTPS server required and verified client cert"`
    ServerCertFile string `json:"server_cert_file" value:"" usage:"Server public key file(*.crt). If both server_cert_file and server_key_file are set, it will set up an HTTPS server"`
    ServerKeyFile  string `json:"server_key_file" value:"" usage:"Server private key file(*.key). If both server_cert_file and server_key_file are set, it will set up an HTTPS server"`
}
```

