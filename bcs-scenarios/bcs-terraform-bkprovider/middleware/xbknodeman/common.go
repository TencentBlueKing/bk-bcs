package xbknodeman

const (
	ServiceName   = "bk-nodeman"
	ServiceNameSg = "nodeman"
	UrlPrefix     = "/prod/api"

	UrlPrefixSgEsb = "api"

	EnvSg = "sg"
)

const (
	// AuthTypePassword  密码认证
	AuthTypePassword = "PASSWORD"
	// AuthTypeKey  秘钥认证
	AuthTypeKey = "KEY"
)

const (
	// EnvBkNodeManHost env for bk-nodeman host
	EnvBkNodeManHost = "bkNodeManHost"
)

// ProxyHost proxy host info
type ProxyHost struct {
	BkCloudId int64  `json:"bk_cloud_id"`
	BkHostId  int64  `json:"bk_host_id"`
	BkBizId   int64  `json:"bk_biz_id"`
	InnerIp   string `json:"inner_ip"`
	InnerIpv6 string `json:"inner_ipv6"`
	OuterIp   string `json:"outer_ip"`
	OuterIpv6 string `json:"outer_ipv6"`
	LoginIp   string `json:"login_ip"`
	DataIp    string `json:"data_ip"`
	ApId      int64  `json:"ap_id"`
	ApName    string `json:"ap_name"`
	Status    string `json:"status"`
	Version   string `json:"version"`
	Port      int64  `json:"port"`
}

// InstallHost install host
type InstallHost struct {
	BkCloudId int64  `json:"bk_cloud_id"`
	BkBizId   int64  `json:"bk_biz_id"`
	BkHostID  int64  `json:"bk_host_id,omitempty"`
	OsType    string `json:"os_type"` // 操作系统，1：LINUX 2：WINDOWS 3：AIX 4：SOLARIS
	InnerIp   string `json:"inner_ip"`
	OuterIp   string `json:"outer_ip,omitempty"`
	LoginIp   string `json:"login_ip,omitempty"`
	Account   string `json:"account,omitempty"`
	Port      int64  `json:"port,omitempty"`
	AuthType  string `json:"auth_type,omitempty"` // 认证类型，1：PASSWORD，密码认证 2: KEY，秘钥认证 3：TJJ_PASSWORD，默认为密码认证
	Password  string `json:"password,omitempty"`
	ApId      int64  `json:"ap_id,omitempty"` // 接入点ID
	Key       string `json:"key,omitempty"`   // 秘钥
}

// Job job
type Job struct {
	JobId  int64  `json:"job_id"`
	JobUrl string `json:"job_url"`
}

// CloudID 	cloud id
type CloudID struct {
	BkCloudID int64 `json:"bk_cloud_id"`
}

// Cloud cloud
type Cloud struct {
	BkCloudId   int64  `json:"bk_cloud_id"`
	BkCloudName string `json:"bk_cloud_name"`
	Isp         string `json:"isp"`
	ApId        int64  `json:"ap_id"`
	IsVisible   bool   `json:"is_visible"`
	NodeCount   int64  `json:"node_count"`
	ProxyCount  int64  `json:"proxy_count"`
	ApName      string `json:"ap_name"`
	IspName     string `json:"isp_name"`
	// IspIcon     string            `json:"isp_icon"` // base64 too big
	Exception   string            `json:"exception"`
	Proxies     []*CloudProxy     `json:"proxies"`
	Permissions *CloudPermissions `json:"permissions"`
}

// CloudProxy proxy
type CloudProxy struct {
	BkCloudId int64  `json:"bk_cloud_id"`
	InnerIp   string `json:"inner_ip"`
	InnerIpv6 string `json:"inner_ipv_6"`
	OuterIp   string `json:"outer_ip"`
	OuterIpv6 string `json:"outer_ipv_6"`
	BkHostId  int64  `json:"bk_host_id"`
	BkAgentId string `json:"bk_agent_id"`
}

// CloudPermissions permissions
type CloudPermissions struct {
	View   bool `json:"view"`
	Edit   bool `json:"edit"`
	Delete bool `json:"delete"`
}

// BizProxyHost biz host
type BizProxyHost struct {
	BkCloudId    int64  `json:"bk_cloud_id"`
	BkAddressing string `json:"bk_addressing"`
	InnerIp      string `json:"inner_ip"`
	InnerIpv6    string `json:"inner_ipv_6"`
	OuterIp      string `json:"outer_ip"`
	OuterIpv6    string `json:"outer_ipv_6"`
	LoginIp      string `json:"login_ip"`
	DataIp       string `json:"data_ip"`
	BkBizId      int64  `json:"bk_biz_id"`
}

// HostInfo host info
type HostInfo struct {
	BkCloudID          int    `json:"bk_cloud_id"`                    // 云区域ID
	BkBizID            int    `json:"bk_biz_id"`                      // 业务ID
	BkHostID           int    `json:"bk_host_id"`                     // 主机ID
	BkHostName         string `json:"bk_host_name"`                   // 主机名
	BkAddressing       string `json:"bk_addressing"`                  // 寻址方式，1: 静态 2: 动态
	OsType             string `json:"os_type"`                        // 操作系统，1：LINUX 2：WINDOWS 3：AIX 4：SOLARIS
	InnerIP            string `json:"inner_ip"`                       // 内网IPv4地址
	InnerIPv6          string `json:"inner_ipv6,omitempty"`           // 内网IPv6地址
	OuterIP            string `json:"outer_ip,omitempty"`             // 外网IPv4地址
	OuterIPv6          string `json:"outer_ipv6,omitempty"`           // 外网IPv6地址
	ApID               int    `json:"ap_id"`                          // 接入点ID
	InstallChannelID   int    `json:"install_channel_id,omitempty"`   // 安装通道ID
	LoginIP            string `json:"login_ip"`                       // 登录IP
	DataIP             string `json:"data_ip"`                        // 数据IP
	Status             string `json:"status"`                         // 运行状态
	Version            string `json:"version"`                        // Agent版本
	CreatedAt          string `json:"created_at"`                     // 创建时间
	UpdatedAt          string `json:"updated_at"`                     // 更新时间
	IsManual           bool   `json:"is_manual"`                      // 是否手动模式
	StatusDisplay      string `json:"status_display,omitempty"`       // 运行执行状态名称
	BkCloudName        string `json:"bk_cloud_name,omitempty"`        // 云区域名称
	InstallChannelName string `json:"install_channel_name,omitempty"` // 安装通道名称
	BkBizName          string `json:"bk_biz_name,omitempty"`          // 业务名称
	OperatePermission  bool   `json:"operate_permission,omitempty"`   // 是否具有操作权限
}
