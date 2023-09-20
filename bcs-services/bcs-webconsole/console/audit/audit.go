package audit

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
)

var (
	auditClient *audit.Client
)

// GetAuditClient 获取审计客户端
func GetAuditClient() *audit.Client {
	if auditClient == nil {
		auditClient =
			audit.NewClient(config.G.BCS.Host, config.G.BCS.Token, nil)
	}
	return auditClient
}
